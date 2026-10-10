package http

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	apitesting "github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/repository"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/service"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/auth"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/resume"
)

type mockRepo struct {
	sessions    map[uuid.UUID]*domain.Session
	items       map[uuid.UUID]*domain.SessionItem
	tasks       map[uuid.UUID]*domain.Task
	answers     map[uuid.UUID]*domain.TestAnswer
	fspProfiles map[uuid.UUID]*domain.CandidateFSPProfile
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		sessions:    make(map[uuid.UUID]*domain.Session),
		items:       make(map[uuid.UUID]*domain.SessionItem),
		tasks:       make(map[uuid.UUID]*domain.Task),
		answers:     make(map[uuid.UUID]*domain.TestAnswer),
		fspProfiles: make(map[uuid.UUID]*domain.CandidateFSPProfile),
	}
}

func (m *mockRepo) GetTemplateForCategory(ctx context.Context, categoryID uuid.UUID) (*domain.Template, error) {
	return &domain.Template{
		ID:         uuid.New(),
		CategoryID: categoryID,
		Config:     json.RawMessage(`{"items_per_session": 1}`),
		IsActive:   true,
	}, nil
}

func (m *mockRepo) GetTasksForTemplate(ctx context.Context, templateID uuid.UUID, limit int) ([]domain.Task, error) {
	t := domain.Task{
		ID:         uuid.New(),
		Type:       domain.TaskTypeSingleChoice,
		Topic:      "go",
		Title:      "Go task",
		Body:       "Sample body",
		Difficulty: 2,
		Solution: json.RawMessage(`{
			"options": [{"id": "opt1", "text": "Answer"}],
			"correct": ["opt1"]
		}`),
	}
	m.tasks[t.ID] = &t
	return []domain.Task{t}, nil
}

func (m *mockRepo) CreateSession(ctx context.Context, session *domain.Session, items []domain.SessionItem) error {
	m.sessions[session.ID] = session
	for _, it := range items {
		itCopy := it
		m.items[it.ID] = &itCopy
	}
	return nil
}

func (m *mockRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	sess, ok := m.sessions[id]
	if !ok {
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

func (m *mockRepo) ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	var list []domain.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockRepo) GetSessionItemWithTask(ctx context.Context, sessionID, itemID uuid.UUID) (*domain.SessionItem, *domain.Task, error) {
	it, ok := m.items[itemID]
	if !ok {
		return nil, nil, repository.ErrNotFound
	}
	task := m.tasks[it.TaskID]
	it.Task = task
	return it, task, nil
}

func (m *mockRepo) SaveAnswer(ctx context.Context, answer *domain.TestAnswer, status domain.SessionItemStatus) error {
	m.answers[answer.ItemID] = answer
	if it, ok := m.items[answer.ItemID]; ok {
		it.Status = status
	}
	return nil
}

func (m *mockRepo) UpdateSessionResult(ctx context.Context, session *domain.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *mockRepo) GetSessionEvaluations(ctx context.Context, sessionID uuid.UUID) ([]repository.SessionItemEval, error) {
	var evals []repository.SessionItemEval
	for _, it := range m.items {
		if it.SessionID == sessionID {
			task := m.tasks[it.TaskID]
			ans := m.answers[it.ID]
			sc := 1.0
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

func (m *mockRepo) GetLatestGradeChange(ctx context.Context, userID uuid.UUID) (*domain.GradeChangeEvent, error) {
	return nil, nil
}

func (m *mockRepo) ListGradeChanges(ctx context.Context, userID uuid.UUID) ([]domain.GradeChangeEvent, error) {
	return nil, nil
}

func (m *mockRepo) CreateGradeChangeEvent(ctx context.Context, event *domain.GradeChangeEvent) error {
	return nil
}

func (m *mockRepo) ListPeriodicTasks(ctx context.Context, categoryID *uuid.UUID) ([]domain.PeriodicTask, error) {
	return nil, nil
}

func (m *mockRepo) CreatePeriodicTask(ctx context.Context, task *domain.PeriodicTask) error {
	return nil
}

func (m *mockRepo) CreatePeriodicSubmission(ctx context.Context, sub *domain.PeriodicSubmission) error {
	return nil
}

func (m *mockRepo) GetCandidateCategory(ctx context.Context, userID uuid.UUID) (*uuid.UUID, *uuid.UUID, *float64, error) {
	return nil, nil, nil, nil
}

func (m *mockRepo) GetGradeDefinition(ctx context.Context, gradeID uuid.UUID) (*domain.GradeDefinition, error) {
	return &domain.GradeDefinition{ID: gradeID, Code: "middle", Name: "Middle", Rank: 4}, nil
}

func (m *mockRepo) GetGradeByRank(ctx context.Context, rank int) (*domain.GradeDefinition, error) {
	return &domain.GradeDefinition{ID: uuid.New(), Code: "middle", Name: "Middle", Rank: rank}, nil
}

func (m *mockRepo) FindCategory(ctx context.Context, specializationID, gradeID uuid.UUID) (*domain.CategoryDefinition, error) {
	return &domain.CategoryDefinition{ID: uuid.New(), SpecializationID: specializationID, GradeID: gradeID, Slug: "cat_test", IsActive: true}, nil
}

func (m *mockRepo) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*domain.CategoryDefinition, error) {
	return &domain.CategoryDefinition{ID: categoryID, SpecializationID: uuid.New(), GradeID: uuid.New(), Slug: "cat_test", IsActive: true}, nil
}

func (m *mockRepo) SaveQuestionnaire(ctx context.Context, q *domain.CandidateQuestionnaire) error {
	return nil
}

func (m *mockRepo) GetLatestQuestionnaire(ctx context.Context, userID uuid.UUID) (*domain.CandidateQuestionnaire, error) {
	return nil, nil
}

func (m *mockRepo) SaveCandidateCategoryState(ctx context.Context, state *domain.CandidateCategoryState) error {
	return nil
}

func (m *mockRepo) GetCandidateCategoryState(ctx context.Context, userID uuid.UUID) (*domain.CandidateCategoryState, error) {
	return nil, nil
}

func (m *mockRepo) SaveCandidateFSP(ctx context.Context, profile *domain.CandidateFSPProfile) error {
	m.fspProfiles[profile.UserID] = profile
	return nil
}

func (m *mockRepo) GetCandidateFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error) {
	if p, ok := m.fspProfiles[userID]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockRepo) UnlinkCandidateFSP(ctx context.Context, userID uuid.UUID) error {
	delete(m.fspProfiles, userID)
	return nil
}

func setupTestServer() (*gin.Engine, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	repo := newMockRepo()
	cfg := &config.Config{
		GradeChangeCooldownDays: 90,
		ItemsPerSession:         1,
		PassThresholdRatio:      0.7,
		UpgradeThresholdRatio:   0.9,
		JWTSecret:               "test-secret-at-least-32-chars-long",
		JWTIssuer:               "fsp-platform",
	}

	svc := service.NewTestingService(repo, cfg)
	handler := NewHandler(svc)

	router := gin.New()
	jwtValidator := auth.NewJWTValidator(cfg.JWTSecret, cfg.JWTIssuer)
	router.Use(auth.Middleware(jwtValidator, true)) // dev mode allows X-User-ID
	apitesting.RegisterHandlers(router, handler)

	// BE2-06 custom routes
	router.POST("/resumes/parse-pdf", handler.ParseResumePDF)
	router.POST("/me/resumes/upload", handler.ParseResumePDF)
	router.POST("/candidates/export-pdf", handler.ExportCandidatePDF)

	testUserID := uuid.New()
	return router, testUserID
}

func TestHTTP_Healthz(t *testing.T) {
	router, _ := setupTestServer()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHTTP_SessionLifecycle(t *testing.T) {
	router, userID := setupTestServer()

	// 1. POST /me/sessions
	startReqBody, _ := json.Marshal(apitesting.StartSessionRequest{
		TargetCategoryId: uuid.New(),
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/me/sessions", bytes.NewReader(startReqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on start session, got %d: %s", w.Code, w.Body.String())
	}

	var session apitesting.Session
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatalf("failed to decode created session: %v", err)
	}

	// 2. GET /me/sessions/{id}
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/me/sessions/"+session.Id.String(), nil)
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on get session, got %d", w.Code)
	}

	var sessWithItems apitesting.SessionWithItems
	if err := json.Unmarshal(w.Body.Bytes(), &sessWithItems); err != nil {
		t.Fatalf("failed to decode session with items: %v", err)
	}

	if sessWithItems.Items == nil || len(*sessWithItems.Items) != 1 {
		t.Fatalf("expected 1 session item")
	}

	itemID := (*sessWithItems.Items)[0].Id

	// 3. POST /me/sessions/{id}/answers
	ansBody, _ := json.Marshal(apitesting.AnswerInput{
		ItemId: itemID,
		Answer: "A",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/me/sessions/"+session.Id.String()+"/answers", bytes.NewReader(ansBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on answer, got %d: %s", w.Code, w.Body.String())
	}

	// 4. POST /me/sessions/{id}/submit
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/me/sessions/"+session.Id.String()+"/submit", nil)
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on submit, got %d: %s", w.Code, w.Body.String())
	}

	var result apitesting.SessionResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode session result: %v", err)
	}

	if result.Decision != apitesting.UpgradeOffered && result.Decision != apitesting.Confirmed {
		t.Errorf("unexpected decision: %s", result.Decision)
	}
}

func TestHTTP_Questionnaire(t *testing.T) {
	router, userID := setupTestServer()

	// 1. Submit questionnaire
	body, _ := json.Marshal(domain.QuestionnaireInput{
		SpecializationID: uuid.New(),
		ClaimedGradeID:   uuid.New(),
		YearsExperience:  3.5,
		Technologies:     []string{"go", "k8s"},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/me/questionnaire", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on questionnaire submit, got %d: %s", w.Code, w.Body.String())
	}

	var qRes domain.QuestionnaireResult
	if err := json.Unmarshal(w.Body.Bytes(), &qRes); err != nil {
		t.Fatalf("failed to decode questionnaire result: %v", err)
	}
	if !qRes.CanStartTest {
		t.Errorf("expected canStartTest = true")
	}

	// 2. Query questionnaire state
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/me/questionnaire", nil)
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on get questionnaire state, got %d", w.Code)
	}
}

func TestHTTP_ParseResumePDF_Multipart(t *testing.T) {
	router, userID := setupTestServer()

	// 1. Generate a valid PDF using resume generator
	sampleProfile := resume.CandidateExportProfile{
		UserID:             userID,
		FullName:           "Александр Дмитриевич Смирнов",
		Headline:           "Senior Backend Developer (Go)",
		SpecializationName: "Backend",
		GradeName:          "Senior",
		Location:           "Москва",
		TestScore:          95.0,
		HasFSP:             true,
		SportsRank:         "Мастер спорта",
		Skills:             []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes"},
		YearsExperience:    6.0,
		MaskContacts:       false,
		Email:              "alex.smirnov@example.com",
		Telegram:           "@alex_backend",
	}
	pdfBytes, err := resume.GenerateCandidatePDF(sampleProfile)
	if err != nil {
		t.Fatalf("failed to generate PDF for test: %v", err)
	}

	// 2. Prepare multipart form request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(pdfBytes); err != nil {
		t.Fatalf("failed to write file part: %v", err)
	}
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/resumes/parse-pdf", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on parse-pdf multipart, got %d: %s", w.Code, w.Body.String())
	}

	var parsed resume.ParsedResume
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if parsed.FullName == "" {
		t.Errorf("expected non-empty fullName in parsed resume, got: %s", w.Body.String())
	}
	if parsed.SuggestedGrade == "" {
		t.Errorf("expected suggestedGrade to be identified")
	}
}

func TestHTTP_ParseResumePDF_JSON(t *testing.T) {
	router, userID := setupTestServer()

	resumeText := `
Елена Николаевна Васильева
Frontend React Engineer
г. Санкт-Петербург

Email: elena.react@mail.ru
Телефон: +7 (911) 555-44-33
Telegram: @elena_fe
GitHub: github.com/elena-fe

Опыт работы: 4 года
2021 — настоящее время
VK Tech, Frontend Разработчик
Стек: React, TypeScript, Vue, Docker.

Образование:
ИТМО, Компьютерные технологии
`
	reqBody, _ := json.Marshal(map[string]string{
		"text": resumeText,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/me/resumes/upload", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on upload resume JSON, got %d: %s", w.Code, w.Body.String())
	}

	var parsed resume.ParsedResume
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed.FullName != "Елена Николаевна Васильева" {
		t.Errorf("expected 'Елена Николаевна Васильева', got '%s'", parsed.FullName)
	}
	if parsed.Contacts.Telegram != "@elena_fe" {
		t.Errorf("expected telegram '@elena_fe', got '%s'", parsed.Contacts.Telegram)
	}
	if parsed.TotalYearsExperience < 3.0 {
		t.Errorf("expected years exp >= 3.0, got %f", parsed.TotalYearsExperience)
	}
}

func TestHTTP_ExportCandidatePDF(t *testing.T) {
	router, userID := setupTestServer()

	exportReq := resume.CandidateExportProfile{
		UserID:             userID,
		FullName:           "Сергей Павлов",
		Headline:           "Lead Python / Go Architect",
		SpecializationName: "Backend",
		GradeName:          "Lead",
		TestScore:          98.0,
		HasFSP:             true,
		SportsRank:         "ЗМС",
		Skills:             []string{"Go", "Python", "Kubernetes", "PostgreSQL"},
		YearsExperience:    8.0,
		MaskContacts:       true,
	}

	reqBody, _ := json.Marshal(exportReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/candidates/export-pdf", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserID, userID.String())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on export-pdf, got %d: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %s", contentType)
	}

	pdfData := w.Body.Bytes()
	if !bytes.HasPrefix(pdfData, []byte("%PDF-1.4")) {
		t.Errorf("expected PDF header %%PDF-1.4")
	}
}


