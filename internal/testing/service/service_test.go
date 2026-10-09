package service

import (
	"context"
	"encoding/json"
	"testing"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/repository"
	"github.com/google/uuid"
)

type mockRepository struct {
	sessions    map[uuid.UUID]*domain.Session
	items       map[uuid.UUID]*domain.SessionItem
	tasks       map[uuid.UUID]*domain.Task
	answers     map[uuid.UUID]*domain.TestAnswer
	gradeEvents []domain.GradeChangeEvent
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sessions: make(map[uuid.UUID]*domain.Session),
		items:    make(map[uuid.UUID]*domain.SessionItem),
		tasks:    make(map[uuid.UUID]*domain.Task),
		answers:  make(map[uuid.UUID]*domain.TestAnswer),
	}
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
	return nil, nil, nil, nil
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
	targetCategoryID := uuid.New()

	// 1. Start session
	sess, err := svc.StartSession(ctx, userID, targetCategoryID)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	if sess.Status != domain.SessionStatusInProgress {
		t.Errorf("expected session to be in_progress, got %s", sess.Status)
	}
	if len(sess.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(sess.Items))
	}

	// 2. Answer items correctly
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

	// 3. Submit session
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

	// 4. Verify cooldown takes effect on immediate next start
	_, err = svc.StartSession(ctx, userID, targetCategoryID)
	if err == nil {
		t.Errorf("expected cooldown error on immediate retake, got nil")
	}
}
