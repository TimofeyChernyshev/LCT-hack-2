package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	repo      repository.Repository
	generator *engine.TaskGenerator
	grader    *engine.Grader
	scorer    *engine.IRTScorer
	anticheat *engine.AntiCheatEngine
	cfg       *config.Config
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
		cfg: cfg,
	}
}

func (s *TestingService) StartSession(ctx context.Context, userID, targetCategoryID uuid.UUID) (*domain.Session, error) {
	// 1. Check cooldown from previous grade changes
	latestChange, err := s.repo.GetLatestGradeChange(ctx, userID)
	if err == nil && latestChange != nil {
		cooldownDuration := time.Duration(s.cfg.GradeChangeCooldownDays) * 24 * time.Hour
		if time.Since(latestChange.ChangedAt) < cooldownDuration {
			// User is within cooldown
			return nil, fmt.Errorf("%w: next change possible after %s",
				ErrCooldownActive, latestChange.ChangedAt.Add(cooldownDuration).Format(time.RFC3339))
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

	// Resulting category is currently the target category
	session.ResultingCategoryID = &session.TargetCategoryID
	// For MVP grade assignment, use target category as resulting grade reference
	resultingGradeID := session.TargetCategoryID
	session.ResultingGradeID = &resultingGradeID

	if err := s.repo.UpdateSessionResult(ctx, session); err != nil {
		return nil, fmt.Errorf("update session evaluated: %w", err)
	}

	// Record grade change event
	event := domain.GradeChangeEvent{
		ID:          uuid.New(),
		UserID:      userID,
		ToGradeID:   resultingGradeID,
		Reason:      fmt.Sprintf("test_session_%s:%s", decision, sessionID.String()),
		ChangedAt:   now,
	}
	_ = s.repo.CreateGradeChangeEvent(ctx, &event)

	return &SessionResultPayload{
		Score:               score,
		AbilityEstimate:     theta,
		ResultingGradeID:    resultingGradeID,
		ResultingCategoryID: session.TargetCategoryID,
		Decision:            decision,
	}, nil
}

func (s *TestingService) ListGradeChanges(ctx context.Context, userID uuid.UUID) (*time.Time, []domain.GradeChangeEvent, error) {
	changes, err := s.repo.ListGradeChanges(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var canChangeAt *time.Time
	if len(changes) > 0 {
		latest := changes[0]
		t := latest.ChangedAt.Add(time.Duration(s.cfg.GradeChangeCooldownDays) * 24 * time.Hour)
		canChangeAt = &t
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
