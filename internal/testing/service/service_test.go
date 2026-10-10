package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/repository"
	"github.com/google/uuid"
)

type mockRepository struct {
	sessions       map[uuid.UUID]*domain.Session
	items          map[uuid.UUID]*domain.SessionItem
	tasks          map[uuid.UUID]*domain.Task
	answers        map[uuid.UUID]*domain.TestAnswer
	gradeEvents    []domain.GradeChangeEvent
	grades         map[uuid.UUID]*domain.GradeDefinition
	gradesByRank   map[int]*domain.GradeDefinition
	categories     map[uuid.UUID]*domain.CategoryDefinition
	categoryStates map[uuid.UUID]*domain.CandidateCategoryState
	questionnaires map[uuid.UUID]*domain.CandidateQuestionnaire
	fspProfiles    map[uuid.UUID]*domain.CandidateFSPProfile
}

func newMockRepository() *mockRepository {
	m := &mockRepository{
		sessions:       make(map[uuid.UUID]*domain.Session),
		items:          make(map[uuid.UUID]*domain.SessionItem),
		tasks:          make(map[uuid.UUID]*domain.Task),
		answers:        make(map[uuid.UUID]*domain.TestAnswer),
		grades:         make(map[uuid.UUID]*domain.GradeDefinition),
		gradesByRank:   make(map[int]*domain.GradeDefinition),
		categories:     make(map[uuid.UUID]*domain.CategoryDefinition),
		categoryStates: make(map[uuid.UUID]*domain.CandidateCategoryState),
		questionnaires: make(map[uuid.UUID]*domain.CandidateQuestionnaire),
		fspProfiles:    make(map[uuid.UUID]*domain.CandidateFSPProfile),
	}

	// Seed grades: 1: intern .. 7: lead
	names := []string{"intern", "junior", "junior_plus", "middle", "middle_plus", "senior", "lead"}
	for i, name := range names {
		g := &domain.GradeDefinition{
			ID:   uuid.New(),
			Code: name,
			Name: name,
			Rank: i + 1,
		}
		m.grades[g.ID] = g
		m.gradesByRank[g.Rank] = g
	}

	return m
}

func (m *mockRepository) GetTemplateForCategory(ctx context.Context, categoryID uuid.UUID) (*domain.Template, error) {
	return &domain.Template{
		ID:         uuid.New(),
		CategoryID: categoryID,
		Config:     json.RawMessage(`{"items_per_session": 2}`),
		IsActive:   true,
	}, nil
}

func (m *mockRepository) GetTasksForTemplate(ctx context.Context, templateID uuid.UUID, limit int) ([]domain.Task, error) {
	t1 := domain.Task{
		ID:         uuid.New(),
		Type:       domain.TaskTypeSingleChoice,
		Topic:      "algorithms",
		Title:      "Question 1",
		Body:       "Body 1",
		Difficulty: 2,
		Solution: json.RawMessage(`{
			"options": [{"id": "opt1", "text": "Opt 1"}, {"id": "opt2", "text": "Opt 2"}],
			"correct": ["opt1"]
		}`),
	}
	t2 := domain.Task{
		ID:         uuid.New(),
		Type:       domain.TaskTypeText,
		Topic:      "math",
		Title:      "Question 2",
		Body:       "Calculate {{v1}}",
		Difficulty: 3,
		Generator: json.RawMessage(`{
			"variables": {"v1": {"type": "int_range", "min": 5, "max": 10}},
			"formula": "v1"
		}`),
	}
	m.tasks[t1.ID] = &t1
	m.tasks[t2.ID] = &t2
	return []domain.Task{t1, t2}, nil
}

func (m *mockRepository) CreateSession(ctx context.Context, session *domain.Session, items []domain.SessionItem) error {
	m.sessions[session.ID] = session
	for _, it := range items {
		itCopy := it
		m.items[it.ID] = &itCopy
	}
	return nil
}

func (m *mockRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	sess, exists := m.sessions[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	sess.Items = nil
	for _, it := range m.items {
		if it.SessionID == id {
			sess.Items = append(sess.Items, *it)
		}
	}
	return sess, nil
}

func (m *mockRepository) ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	var list []domain.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockRepository) GetSessionItemWithTask(ctx context.Context, sessionID, itemID uuid.UUID) (*domain.SessionItem, *domain.Task, error) {
	it, exists := m.items[itemID]
	if !exists {
		return nil, nil, repository.ErrNotFound
	}
	task := m.tasks[it.TaskID]
	it.Task = task
	return it, task, nil
}

func (m *mockRepository) SaveAnswer(ctx context.Context, answer *domain.TestAnswer, status domain.SessionItemStatus) error {
	m.answers[answer.ItemID] = answer
	if it, exists := m.items[answer.ItemID]; exists {
		it.Status = status
	}
	return nil
}

func (m *mockRepository) UpdateSessionResult(ctx context.Context, session *domain.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockRepository) GetSessionEvaluations(ctx context.Context, sessionID uuid.UUID) ([]repository.SessionItemEval, error) {
	var evals []repository.SessionItemEval
	for _, it := range m.items {
		if it.SessionID == sessionID {
			task := m.tasks[it.TaskID]
			ans := m.answers[it.ID]
			sc := 0.0
			if ans != nil && ans.Score != nil {
				sc = *ans.Score
			}
			evals = append(evals, repository.SessionItemEval{
				Item:           *it,
				Task:           *task,
				Answer:         ans,
				Discrimination: 1.0,
				Difficulty:     task.Difficulty,
				Score:          sc,
			})
		}
	}
	return evals, nil
}

func (m *mockRepository) GetLatestGradeChange(ctx context.Context, userID uuid.UUID) (*domain.GradeChangeEvent, error) {
	if len(m.gradeEvents) == 0 {
		return nil, nil
	}
	return &m.gradeEvents[len(m.gradeEvents)-1], nil
}

func (m *mockRepository) ListGradeChanges(ctx context.Context, userID uuid.UUID) ([]domain.GradeChangeEvent, error) {
	return m.gradeEvents, nil
}

func (m *mockRepository) CreateGradeChangeEvent(ctx context.Context, event *domain.GradeChangeEvent) error {
	m.gradeEvents = append(m.gradeEvents, *event)
	return nil
}

func (m *mockRepository) ListPeriodicTasks(ctx context.Context, categoryID *uuid.UUID) ([]domain.PeriodicTask, error) {
	return nil, nil
}

func (m *mockRepository) CreatePeriodicTask(ctx context.Context, task *domain.PeriodicTask) error {
	return nil
}

func (m *mockRepository) CreatePeriodicSubmission(ctx context.Context, sub *domain.PeriodicSubmission) error {
	return nil
}

func (m *mockRepository) GetCandidateCategory(ctx context.Context, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, *float64, error) {
	if st, ok := m.categoryStates[userID]; ok {
		return &st.CurrentCategoryID, &st.CurrentGradeID, &st.TestScore, nil
	}
	return nil, nil, nil, nil
}

func (m *mockRepository) GetGradeDefinition(ctx context.Context, gradeID uuid.UUID) (*domain.GradeDefinition, error) {
	if g, ok := m.grades[gradeID]; ok {
		return g, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepository) GetGradeByRank(ctx context.Context, rank int) (*domain.GradeDefinition, error) {
	if g, ok := m.gradesByRank[rank]; ok {
		return g, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepository) FindCategory(ctx context.Context, specializationID, gradeID uuid.UUID) (*domain.CategoryDefinition, error) {
	for _, c := range m.categories {
		if c.SpecializationID == specializationID && c.GradeID == gradeID {
			return c, nil
		}
	}
	c := &domain.CategoryDefinition{
		ID:               uuid.New(),
		SpecializationID: specializationID,
		GradeID:          gradeID,
		Slug:             "test_cat",
		IsActive:         true,
	}
	m.categories[c.ID] = c
	return c, nil
}

func (m *mockRepository) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*domain.CategoryDefinition, error) {
	if c, ok := m.categories[categoryID]; ok {
		return c, nil
	}
	return nil, repository.ErrNotFound
}

func (m *mockRepository) SaveQuestionnaire(ctx context.Context, q *domain.CandidateQuestionnaire) error {
	m.questionnaires[q.UserID] = q
	return nil
}

func (m *mockRepository) GetLatestQuestionnaire(ctx context.Context, userID uuid.UUID) (*domain.CandidateQuestionnaire, error) {
	if q, ok := m.questionnaires[userID]; ok {
		return q, nil
	}
	return nil, nil
}

func (m *mockRepository) SaveCandidateCategoryState(ctx context.Context, state *domain.CandidateCategoryState) error {
	m.categoryStates[state.UserID] = state
	return nil
}

func (m *mockRepository) GetCandidateCategoryState(ctx context.Context, userID uuid.UUID) (*domain.CandidateCategoryState, error) {
	if s, ok := m.categoryStates[userID]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockRepository) SaveCandidateFSP(ctx context.Context, profile *domain.CandidateFSPProfile) error {
	m.fspProfiles[profile.UserID] = profile
	return nil
}

func (m *mockRepository) GetCandidateFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error) {
	if p, ok := m.fspProfiles[userID]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockRepository) UnlinkCandidateFSP(ctx context.Context, userID uuid.UUID) error {
	delete(m.fspProfiles, userID)
	return nil
}

func TestTestingService_CompleteFlow(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		GradeChangeCooldownDays: 90,
		ItemsPerSession:         2,
		PassThresholdRatio:      0.7,
		UpgradeThresholdRatio:   0.9,
	}

	svc := NewTestingService(repo, cfg)
	ctx := context.Background()
	userID := uuid.New()

	// Junior grade (rank 2)
	juniorGrade := repo.gradesByRank[2]
	specID := uuid.New()

	// 1. Questionnaire onboarding
	qRes, err := svc.EvaluateQuestionnaire(ctx, userID, domain.QuestionnaireInput{
		SpecializationID: specID,
		ClaimedGradeID:   juniorGrade.ID,
		YearsExperience:  1.0,
		Technologies:     []string{"go", "postgres"},
	})
	if err != nil {
		t.Fatalf("failed questionnaire: %v", err)
	}
	if !qRes.CanStartTest {
		t.Fatalf("expected canStartTest = true")
	}

	// 2. Start session
	sess, err := svc.StartSession(ctx, userID, qRes.TargetCategoryID)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// 3. Answer items 100% correctly
	for _, it := range sess.Items {
		var variant struct {
			CorrectKeys   []string `json:"correct_keys"`
			ExpectedValue string   `json:"expected"`
		}
		_ = json.Unmarshal(it.VariantParams, &variant)

		var answer interface{}
		if len(variant.CorrectKeys) > 0 {
			answer = variant.CorrectKeys[0]
		} else {
			answer = variant.ExpectedValue
		}

		err = svc.AnswerItem(ctx, userID, sess.ID, it.ID, answer, false)
		if err != nil {
			t.Fatalf("failed to answer item: %v", err)
		}
	}

	// 4. Submit session
	res, err := svc.SubmitSession(ctx, userID, sess.ID)
	if err != nil {
		t.Fatalf("failed to submit session: %v", err)
	}

	if res.Score != 100.0 {
		t.Errorf("expected 100 score, got %v", res.Score)
	}
	if res.Decision != "upgrade_offered" {
		t.Errorf("expected upgrade_offered, got %s", res.Decision)
	}

	// Resulting grade should be promoted to rank 3 (junior_plus)
	promotedGrade, _ := repo.GetGradeDefinition(ctx, res.ResultingGradeID)
	if promotedGrade == nil || promotedGrade.Rank != 3 {
		t.Errorf("expected promoted rank 3, got %v", promotedGrade)
	}

	// 5. Cooldown check: Immediate retake should be blocked
	_, err = svc.StartSession(ctx, userID, qRes.TargetCategoryID)
	if err == nil {
		t.Errorf("expected cooldown error on immediate retake, got nil")
	}

	// 6. Candidate category state check
	catID, gradeID, score, err := svc.GetCandidateCategory(ctx, userID)
	if err != nil || catID == nil || gradeID == nil || score == nil {
		t.Fatalf("expected candidate category to be set in state")
	}
	if *gradeID != res.ResultingGradeID {
		t.Errorf("expected candidate state grade to match resultingGradeID")
	}
}

func TestTestingService_DowngradePreservesBaseGrade(t *testing.T) {
	repo := newMockRepository()
	cfg := &config.Config{
		GradeChangeCooldownDays: 90,
		ItemsPerSession:         2,
		PassThresholdRatio:      0.7,
		UpgradeThresholdRatio:   0.9,
	}

	svc := NewTestingService(repo, cfg)
	ctx := context.Background()
	userID := uuid.New()

	// Suppose user already had confirmed Middle grade (rank 4)
	middleGrade := repo.gradesByRank[4]
	seniorGrade := repo.gradesByRank[6]
	specID := uuid.New()

	middleCat, _ := repo.FindCategory(ctx, specID, middleGrade.ID)
	seniorCat, _ := repo.FindCategory(ctx, specID, seniorGrade.ID)

	pastSessionID := uuid.New()
	pastTime := time.Now().Add(-100 * 24 * time.Hour) // cooldown expired
	_ = repo.SaveCandidateCategoryState(ctx, &domain.CandidateCategoryState{
		UserID:            userID,
		CurrentCategoryID: middleCat.ID,
		CurrentGradeID:    middleGrade.ID,
		Status:            "confirmed",
		TestScore:         80.0,
		LastSessionID:     pastSessionID,
		UpdatedAt:         pastTime,
	})

	// User now attempts test for Senior (rank 6)
	sess, err := svc.StartSession(ctx, userID, seniorCat.ID)
	if err != nil {
		t.Fatalf("failed to start senior session: %v", err)
	}

	// User fails test (wrong answers -> 0%)
	for _, it := range sess.Items {
		_ = svc.AnswerItem(ctx, userID, sess.ID, it.ID, "WRONG_ANSWER", false)
	}

	// Submit session
	res, err := svc.SubmitSession(ctx, userID, sess.ID)
	if err != nil {
		t.Fatalf("failed to submit session: %v", err)
	}

	if res.Decision != "downgrade_offered" {
		t.Errorf("expected downgrade_offered, got %s", res.Decision)
	}

	// ТЗ DoD: При неуспешном прохождении грейд не срезается принудительно «в ноль».
	// User had Middle (rank 4). Since rank 4 >= lower recommendation (rank 5), user preserves rank 4!
	resultingGrade, _ := repo.GetGradeDefinition(ctx, res.ResultingGradeID)
	if resultingGrade == nil || resultingGrade.Rank < 4 {
		t.Errorf("expected base confirmed grade (rank >= 4) not to be dropped to 0, got rank %v", resultingGrade)
	}
}
