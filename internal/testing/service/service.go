package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/engine"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/repository"
)

var (
	ErrSessionNotFound     = errors.New("test session not found")
	ErrSessionClosed       = errors.New("test session is closed or expired")
	ErrItemAlreadyAnswered = errors.New("test item already answered or session closed")
	ErrCooldownActive      = errors.New("grade change cooldown is active")
	ErrUnauthorized        = errors.New("unauthorized to access session")
	ErrNoTasksAvailable    = errors.New("no tasks available for category")
)

type TestingService struct {
	repo       repository.Repository
	generator  *engine.TaskGenerator
	grader     *engine.Grader
	scorer     *engine.IRTScorer
	anticheat  *engine.AntiCheatEngine
	cfg        *config.Config
	httpClient *http.Client
}

func NewTestingService(repo repository.Repository, cfg *config.Config) *TestingService {
	sessionDuration := 45 * time.Minute
	return &TestingService{
		repo:      repo,
		generator: engine.NewTaskGenerator(),
		grader:    engine.NewGrader(),
		scorer:    engine.NewIRTScorer(cfg.PassThresholdRatio, cfg.UpgradeThresholdRatio),
		anticheat: engine.NewAntiCheatEngine(engine.AntiCheatConfig{
			SessionDuration:      sessionDuration,
			MinAnswerTimeSeconds: 2,
		}),
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.HTTPTimeout},
	}
}

// EvaluateQuestionnaire handles pre-test onboarding questionnaire, claiming specialization/grade and checking cooldown
func (s *TestingService) EvaluateQuestionnaire(ctx context.Context, userID uuid.UUID, input domain.QuestionnaireInput) (*domain.QuestionnaireResult, error) {
	// 1. Check current cooldown and status
	state, err := s.repo.GetCandidateCategoryState(ctx, userID)
	if err == nil && state != nil && state.CanChangeAt != nil {
		if time.Now().Before(*state.CanChangeAt) {
			remaining := int(time.Until(*state.CanChangeAt).Hours()/24) + 1

			// Check smart retry: if candidate was offered downgrade and is retaking on recommended lower grade, allow it
			isLowerRetake := false
			if state.Status == "downgrade_offered" {
				claimedGrade, _ := s.repo.GetGradeDefinition(ctx, input.ClaimedGradeID)
				currentGrade, _ := s.repo.GetGradeDefinition(ctx, state.CurrentGradeID)
				if claimedGrade != nil && currentGrade != nil && claimedGrade.Rank <= currentGrade.Rank {
					isLowerRetake = true
				}
			}

			if !isLowerRetake {
				return &domain.QuestionnaireResult{
					TargetCategoryID:      state.CurrentCategoryID,
					RecommendedGradeID:    state.CurrentGradeID,
					CanStartTest:          false,
					CooldownRemainingDays: remaining,
					CanChangeAt:           state.CanChangeAt,
					Message:               fmt.Sprintf("Кулдаун смены грейда активен. Повторная сдача станет доступна через %d дн. (%s).", remaining, state.CanChangeAt.Format("02.01.2006")),
				}, nil
			}
		}
	}

	// 2. Resolve category matching specialization + claimed grade
	targetCategoryID := uuid.Nil
	catDef, err := s.repo.FindCategory(ctx, input.SpecializationID, input.ClaimedGradeID)
	if err == nil && catDef != nil {
		targetCategoryID = catDef.ID
	} else {
		// Fallback to active template category or claimed grade ID as category reference
		tmpl, err := s.repo.GetTemplateForCategory(ctx, input.ClaimedGradeID)
		if err == nil && tmpl != nil {
			targetCategoryID = tmpl.CategoryID
		} else {
			targetCategoryID = input.ClaimedGradeID
		}
	}

	// 3. Experience check & recommendation logic
	recommendedGradeID := input.ClaimedGradeID
	claimedGrade, err := s.repo.GetGradeDefinition(ctx, input.ClaimedGradeID)
	message := "Опросник успешно заполнен. Вы можете приступить к тестированию для подтверждения грейда."

	if err == nil && claimedGrade != nil {
		// Experience guideline: Junior (0-1.5y), Middle (1.5-4y), Senior (4y+)
		if claimedGrade.Rank >= 6 && input.YearsExperience < 2.0 {
			message = fmt.Sprintf("Вы заявили грейд %s при опыте %.1f г. Рекомендуется пройти тест для подтверждения уровня компетенций.", claimedGrade.Name, input.YearsExperience)
		} else if claimedGrade.Rank <= 2 && input.YearsExperience > 4.0 {
			message = fmt.Sprintf("С вашим опытом (%.1f г.) вы можете претендовать на более высокий грейд (Middle/Senior). Тест покажет объективный уровень.", input.YearsExperience)
		}
	}

	// 4. Save questionnaire
	techJSON, _ := json.Marshal(input.Technologies)
	q := domain.CandidateQuestionnaire{
		ID:                 uuid.New(),
		UserID:             userID,
		SpecializationID:   input.SpecializationID,
		ClaimedGradeID:     input.ClaimedGradeID,
		TargetCategoryID:   targetCategoryID,
		YearsExperience:    input.YearsExperience,
		Technologies:       techJSON,
		CreatedAt:          time.Now(),
	}
	_ = s.repo.SaveQuestionnaire(ctx, &q)

	return &domain.QuestionnaireResult{
		TargetCategoryID:      targetCategoryID,
		RecommendedGradeID:    recommendedGradeID,
		CanStartTest:          true,
		CooldownRemainingDays: 0,
		CanChangeAt:           nil,
		Message:               message,
	}, nil
}

// GetQuestionnaireState retrieves current questionnaire and cooldown status for the candidate
func (s *TestingService) GetQuestionnaireState(ctx context.Context, userID uuid.UUID) (*domain.CandidateCategoryState, *time.Time, bool, error) {
	state, err := s.repo.GetCandidateCategoryState(ctx, userID)
	if err != nil {
		return nil, nil, true, err
	}

	if state == nil {
		return nil, nil, true, nil
	}

	canStart := true
	var canChangeAt *time.Time
	if state.CanChangeAt != nil {
		canChangeAt = state.CanChangeAt
		if time.Now().Before(*state.CanChangeAt) {
			canStart = false
		}
	}

	return state, canChangeAt, canStart, nil
}

func (s *TestingService) StartSession(ctx context.Context, userID, targetCategoryID uuid.UUID) (*domain.Session, error) {
	// 1. Check cooldown
	state, err := s.repo.GetCandidateCategoryState(ctx, userID)
	if err == nil && state != nil && state.CanChangeAt != nil {
		if time.Now().Before(*state.CanChangeAt) {
			// Check if candidate is retaking on recommended lower grade after downgrade_offered
			isLowerRetake := false
			if state.Status == "downgrade_offered" {
				catDef, _ := s.repo.GetCategoryByID(ctx, targetCategoryID)
				if catDef != nil {
					targetGrade, _ := s.repo.GetGradeDefinition(ctx, catDef.GradeID)
					currGrade, _ := s.repo.GetGradeDefinition(ctx, state.CurrentGradeID)
					if targetGrade != nil && currGrade != nil && targetGrade.Rank <= currGrade.Rank {
						isLowerRetake = true
					}
				}
			}

			if !isLowerRetake {
				remaining := int(time.Until(*state.CanChangeAt).Hours()/24) + 1
				return nil, fmt.Errorf("%w: повторное тестирование возможно через %d дн. (%s)",
					ErrCooldownActive, remaining, state.CanChangeAt.Format(time.RFC3339))
			}
		}
	}

	// 2. Fetch template for target category
	tmpl, err := s.repo.GetTemplateForCategory(ctx, targetCategoryID)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}

	// Determine items per session from config / template
	itemsCount := s.cfg.ItemsPerSession
	if itemsCount <= 0 {
		itemsCount = 10
	}
	if len(tmpl.Config) > 0 {
		var tc domain.TemplateConfig
		if err := json.Unmarshal(tmpl.Config, &tc); err == nil && tc.ItemsPerSession > 0 {
			itemsCount = tc.ItemsPerSession
		}
	}

	// 3. Select balanced tasks from bank
	tasks, err := s.repo.GetTasksForTemplate(ctx, tmpl.ID, itemsCount)
	if err != nil {
		return nil, fmt.Errorf("get tasks: %w", err)
	}
	if len(tasks) == 0 {
		return nil, ErrNoTasksAvailable
	}

	sessionID := uuid.New()
	session := domain.Session{
		ID:               sessionID,
		UserID:           userID,
		TargetCategoryID: targetCategoryID,
		TemplateID:       tmpl.ID,
		Status:           domain.SessionStatusInProgress,
		StartedAt:        time.Now(),
	}

	// 4. Generate unique variant for each question (Anti-Cheat parameterization + option shuffle)
	items := make([]domain.SessionItem, 0, len(tasks))
	for i, t := range tasks {
		item, err := s.generator.GenerateItem(&t, sessionID.String(), i+1)
		if err != nil {
			return nil, fmt.Errorf("generate item %d: %w", i+1, err)
		}
		item.ID = uuid.New()
		item.SessionID = sessionID
		item.TaskID = t.ID
		items = append(items, *item)
	}

	// 5. Persist session and items in DB
	if err := s.repo.CreateSession(ctx, &session, items); err != nil {
		return nil, fmt.Errorf("create session in db: %w", err)
	}

	session.Items = items
	return &session, nil
}

func (s *TestingService) GetSession(ctx context.Context, userID, sessionID uuid.UUID) (*domain.Session, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if session.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Check timeout
	if session.Status == domain.SessionStatusInProgress {
		if err := s.anticheat.ValidateSessionTime(session); err != nil {
			session.Status = domain.SessionStatusExpired
			now := time.Now()
			session.FinishedAt = &now
			_ = s.repo.UpdateSessionResult(ctx, session)
		}
	}

	return session, nil
}

func (s *TestingService) ListMySessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	return s.repo.ListSessionsByUserID(ctx, userID)
}

func (s *TestingService) AnswerItem(ctx context.Context, userID, sessionID, itemID uuid.UUID, rawAnswer interface{}, skipped bool) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	if session.UserID != userID {
		return ErrUnauthorized
	}

	if session.Status != domain.SessionStatusInProgress {
		return ErrSessionClosed
	}

	// Anti-cheat: Check if session time has expired
	if err := s.anticheat.ValidateSessionTime(session); err != nil {
		session.Status = domain.SessionStatusExpired
		now := time.Now()
		session.FinishedAt = &now
		_ = s.repo.UpdateSessionResult(ctx, session)
		return ErrSessionClosed
	}

	// Fetch item and associated task
	item, task, err := s.repo.GetSessionItemWithTask(ctx, sessionID, itemID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	if item.Status != domain.SessionItemPending {
		return ErrItemAlreadyAnswered
	}

	now := time.Now()
	var answerBytes []byte
	if rawAnswer != nil {
		answerBytes, _ = json.Marshal(rawAnswer)
	} else {
		answerBytes = []byte("null")
	}

	ans := domain.TestAnswer{
		ID:         uuid.New(),
		ItemID:     itemID,
		Answer:     answerBytes,
		GradedBy:   "auto",
		GradedAt:   &now,
		AnsweredAt: now,
	}

	var newStatus domain.SessionItemStatus
	if skipped {
		isCorr := false
		sc := 0.0
		ans.IsCorrect = &isCorr
		ans.Score = &sc
		newStatus = domain.SessionItemSkipped
	} else {
		gradeRes := s.grader.Grade(task, item, rawAnswer)
		ans.IsCorrect = &gradeRes.IsCorrect
		ans.Score = &gradeRes.Score
		newStatus = domain.SessionItemAnswered
	}

	if err := s.repo.SaveAnswer(ctx, &ans, newStatus); err != nil {
		if errors.Is(err, repository.ErrAlreadyAnswer) {
			return ErrItemAlreadyAnswered
		}
		return fmt.Errorf("save answer: %w", err)
	}

	return nil
}

type SessionResultPayload struct {
	Score               float64
	AbilityEstimate     float64
	ResultingGradeID    uuid.UUID
	ResultingCategoryID uuid.UUID
	Decision            string
	Explanation         string
}

func (s *TestingService) SubmitSession(ctx context.Context, userID, sessionID uuid.UUID) (*SessionResultPayload, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if session.UserID != userID {
		return nil, ErrUnauthorized
	}

	if session.Status != domain.SessionStatusInProgress {
		return nil, ErrSessionClosed
	}

	evals, err := s.repo.GetSessionEvaluations(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get evals: %w", err)
	}

	// Transform to IRT item evaluations
	irtItems := make([]engine.ItemEvaluation, 0, len(evals))
	for _, ev := range evals {
		b := engine.MapDifficultyToIRT(ev.Difficulty, ev.DifficultyIRT)
		irtItems = append(irtItems, engine.ItemEvaluation{
			Discrimination: ev.Discrimination,
			Difficulty:     b,
			Score:          ev.Score,
		})
	}

	score, theta, decision := s.scorer.ComputeSessionResult(irtItems)

	now := time.Now()
	session.Status = domain.SessionStatusEvaluated
	session.Score = &score
	session.AbilityEstimate = &theta
	session.FinishedAt = &now

	// 1. Resolve claimed grade & specialization
	var claimedGradeID uuid.UUID
	var specID *uuid.UUID
	claimedRank := 4 // Default Middle

	catDef, _ := s.repo.GetCategoryByID(ctx, session.TargetCategoryID)
	if catDef != nil {
		claimedGradeID = catDef.GradeID
		specID = &catDef.SpecializationID
		gDef, _ := s.repo.GetGradeDefinition(ctx, catDef.GradeID)
		if gDef != nil {
			claimedRank = gDef.Rank
		}
	} else {
		// Fallback to latest questionnaire
		quest, _ := s.repo.GetLatestQuestionnaire(ctx, userID)
		if quest != nil {
			claimedGradeID = quest.ClaimedGradeID
			specID = &quest.SpecializationID
			gDef, _ := s.repo.GetGradeDefinition(ctx, quest.ClaimedGradeID)
			if gDef != nil {
				claimedRank = gDef.Rank
			}
		} else {
			claimedGradeID = session.TargetCategoryID
		}
	}

	// Check previous state
	prevState, _ := s.repo.GetCandidateCategoryState(ctx, userID)

	// 2. BE2-02 Logic for grade progression & decision
	resultingGradeID := claimedGradeID
	resultingCategoryID := session.TargetCategoryID
	var explanation string

	switch decision {
	case "upgrade_offered":
		// Confirmed with high score (>= 90%) -> offer upgrade to next rank (Rank + 1)
		nextRank := claimedRank + 1
		if nextRank > 7 {
			nextRank = 7
		}
		upgradedGrade, err := s.repo.GetGradeByRank(ctx, nextRank)
		if err == nil && upgradedGrade != nil {
			resultingGradeID = upgradedGrade.ID
			if specID != nil {
				upgradedCat, err := s.repo.FindCategory(ctx, *specID, upgradedGrade.ID)
				if err == nil && upgradedCat != nil {
					resultingCategoryID = upgradedCat.ID
				}
			}
			explanation = fmt.Sprintf("Отличный результат (%.1f%%). Заявленный уровень подтвержден с отличием, предложено повышение до %s.", score, upgradedGrade.Name)
		} else {
			explanation = fmt.Sprintf("Отличный результат (%.1f%%). Заявленный грейд подтвержден с отличием.", score)
		}

	case "confirmed":
		// Normal pass (70..89%) -> confirm claimed grade
		claimedGDef, _ := s.repo.GetGradeDefinition(ctx, claimedGradeID)
		gradeName := "заявленного уровня"
		if claimedGDef != nil {
			gradeName = claimedGDef.Name
		}
		explanation = fmt.Sprintf("Успешное прохождение (%.1f%%). Грейд %s официально подтвержден результатами тестирования.", score, gradeName)

	case "downgrade_offered":
		// Failed to reach pass threshold (< 70%)
		// ТЗ: Грейд не срезается принудительно «в ноль», предлагается сдать тест на уровень ниже.
		lowerRank := claimedRank - 1
		if lowerRank < 1 {
			lowerRank = 1
		}
		lowerGrade, err := s.repo.GetGradeByRank(ctx, lowerRank)
		if err == nil && lowerGrade != nil {
			resultingGradeID = lowerGrade.ID
			if specID != nil {
				lowerCat, err := s.repo.FindCategory(ctx, *specID, lowerGrade.ID)
				if err == nil && lowerCat != nil {
					resultingCategoryID = lowerCat.ID
				}
			}

			// If candidate already had an established confirmed grade from past tests, preserve it
			if prevState != nil && prevState.Status == "confirmed" {
				prevGrade, _ := s.repo.GetGradeDefinition(ctx, prevState.CurrentGradeID)
				if prevGrade != nil && prevGrade.Rank >= lowerRank {
					// Base confirmed grade is preserved!
					resultingGradeID = prevState.CurrentGradeID
					resultingCategoryID = prevState.CurrentCategoryID
				}
			}
			explanation = fmt.Sprintf("Балл (%.1f%%) ниже порога подтверждения (70%%). Текущий статус сохранен, рекомендована пересдача на уровень %s.", score, lowerGrade.Name)
		} else {
			explanation = fmt.Sprintf("Балл (%.1f%%) ниже порога подтверждения. Рекомендована подготовка и повторная сдача.", score)
		}
	}

	session.ResultingGradeID = &resultingGradeID
	session.ResultingCategoryID = &resultingCategoryID

	if err := s.repo.UpdateSessionResult(ctx, session); err != nil {
		return nil, fmt.Errorf("update session evaluated: %w", err)
	}

	// 3. Cooldown calculation & Candidate Category State persistence
	cooldownDuration := time.Duration(s.cfg.GradeChangeCooldownDays) * 24 * time.Hour
	canChangeAt := now.Add(cooldownDuration)

	state := domain.CandidateCategoryState{
		UserID:            userID,
		CurrentCategoryID: resultingCategoryID,
		CurrentGradeID:    resultingGradeID,
		SpecializationID:  specID,
		Status:            decision,
		TestScore:         score,
		AbilityEstimate:   &theta,
		LastSessionID:     sessionID,
		CanChangeAt:       &canChangeAt,
		UpdatedAt:         now,
	}
	if err := s.repo.SaveCandidateCategoryState(ctx, &state); err != nil {
		log.Printf("[testing-service] warning: failed to save category state: %v", err)
	}

	// 4. Record grade change event
	var fromGradeID *uuid.UUID
	if prevState != nil {
		fromGradeID = &prevState.CurrentGradeID
	}
	event := domain.GradeChangeEvent{
		ID:          uuid.New(),
		UserID:      userID,
		FromGradeID: fromGradeID,
		ToGradeID:   resultingGradeID,
		Reason:      fmt.Sprintf("test_session_%s:%s (%s)", decision, sessionID.String(), explanation),
		ChangedAt:   now,
	}
	_ = s.repo.CreateGradeChangeEvent(ctx, &event)

	// 5. Notify Candidate Service asynchronously
	go s.notifyCandidateService(userID, resultingCategoryID, resultingGradeID, score, decision)

	return &SessionResultPayload{
		Score:               score,
		AbilityEstimate:     theta,
		ResultingGradeID:    resultingGradeID,
		ResultingCategoryID: resultingCategoryID,
		Decision:            decision,
		Explanation:         explanation,
	}, nil
}

func (s *TestingService) ListGradeChanges(ctx context.Context, userID uuid.UUID) (*time.Time, []domain.GradeChangeEvent, error) {
	changes, err := s.repo.ListGradeChanges(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check canChangeAt from candidate_category_state first
	state, _ := s.repo.GetCandidateCategoryState(ctx, userID)
	if state != nil && state.CanChangeAt != nil {
		if time.Now().Before(*state.CanChangeAt) {
			return state.CanChangeAt, changes, nil
		}
		return nil, changes, nil
	}

	var canChangeAt *time.Time
	if len(changes) > 0 {
		latest := changes[0]
		t := latest.ChangedAt.Add(time.Duration(s.cfg.GradeChangeCooldownDays) * 24 * time.Hour)
		if time.Now().Before(t) {
			canChangeAt = &t
		}
	}

	return canChangeAt, changes, nil
}

func (s *TestingService) GetCandidateCategory(ctx context.Context, userID uuid.UUID) (catID, gradeID *uuid.UUID, score *float64, err error) {
	return s.repo.GetCandidateCategory(ctx, userID)
}

func (s *TestingService) ListPeriodicTasks(ctx context.Context, categoryID *uuid.UUID) ([]domain.PeriodicTask, error) {
	return s.repo.ListPeriodicTasks(ctx, categoryID)
}

func (s *TestingService) CreatePeriodicTask(ctx context.Context, employerUserID, categoryID uuid.UUID, title, body string) (*domain.PeriodicTask, error) {
	task := domain.PeriodicTask{
		ID:             uuid.New(),
		EmployerUserID: employerUserID,
		CategoryID:     categoryID,
		Title:          title,
		Body:           body,
		IsActive:       true,
		CreatedAt:      time.Now(),
	}
	if err := s.repo.CreatePeriodicTask(ctx, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TestingService) SubmitPeriodicTask(ctx context.Context, userID, taskID uuid.UUID, answer string) error {
	sub := domain.PeriodicSubmission{
		ID:        uuid.New(),
		TaskID:    taskID,
		UserID:    userID,
		Answer:    answer,
		CreatedAt: time.Now(),
	}
	sc := 1.0
	sub.Score = &sc
	return s.repo.CreatePeriodicSubmission(ctx, &sub)
}

// notifyCandidateService sends updated category to candidate microservice if reachable
func (s *TestingService) notifyCandidateService(userID, categoryID, gradeID uuid.UUID, score float64, decision string) {
	if s.cfg.CandidateURL == "" {
		return
	}
	url := fmt.Sprintf("%s/internal/candidates/%s/category", s.cfg.CandidateURL, userID.String())
	payload := map[string]interface{}{
		"categoryId": categoryID.String(),
		"gradeId":    gradeID.String(),
		"score":      score,
		"decision":   decision,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Non-fatal if candidate service is not running yet during testing
		return
	}
	_ = resp.Body.Close()
}
