package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	"TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/config"
	apisearch "TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http"
	"TimofeyChernyshev/LCT-hack-2/internal/search/service"
)

type mockRepo struct {
	docs []domain.CandidateSearchDoc
}

func (m *mockRepo) Search(ctx context.Context, filter domain.SearchFilter) ([]domain.CandidateSearchDoc, int, error) {
	return m.docs, len(m.docs), nil
}

func (m *mockRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.CandidateSearchDoc, error) {
	for _, d := range m.docs {
		if d.UserID == userID {
			return &d, nil
		}
	}
	return nil, assert.AnError
}

func (m *mockRepo) GetCategoryRanking(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.CandidateSearchDoc, error) {
	return m.docs, nil
}

func (m *mockRepo) CalculateTestPercentile(ctx context.Context, categoryID uuid.UUID, testScore float64) (float64, error) {
	return 95.0, nil
}

func (m *mockRepo) Upsert(ctx context.Context, doc *domain.CandidateSearchDoc) error {
	m.docs = append(m.docs, *doc)
	return nil
}

func (m *mockRepo) Count(ctx context.Context) (int, error) {
	return len(m.docs), nil
}

func setupTestRouter() (*gin.Engine, *mockRepo, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	candID := uuid.New()
	catID := uuid.New()
	specID := uuid.New()
	gradeID := uuid.New()
	testScore := 95.0
	place := 1

	repo := &mockRepo{
		docs: []domain.CandidateSearchDoc{
			{
				UserID:               candID,
				DisplayName:          "Иван Смирнов",
				CategoryID:           catID,
				SpecializationID:     specID,
				SpecializationName:   "Backend Go",
				GradeID:              gradeID,
				GradeName:            "Senior",
				TestScore:            &testScore,
				HasFSP:               true,
				SportsRank:           "Мастер спорта",
				FSPScore:             90.0,
				FSPBestPlace:         &place,
				FSPWeightSum:         25,
				FSPHighlights:        []string{"1 место Чемпионат ФСП 2025"},
				CalculatedScore:      94.0,
				Explanation:          "Топ-5% по тесту, 1 место ФСП",
				Reasons:              []string{"top_test_performer", "fsp_champion"},
				LastActiveAt:         time.Now(),
			},
		},
	}

	cfg := &config.Config{DefaultLimit: 20, MaxLimit: 100}
	svc := service.NewSearchService(repo, cfg)
	handler := NewHandler(svc)

	apisearch.RegisterHandlers(router, handler)

	return router, repo, candID
}

func TestHandlerHealthz(t *testing.T) {
	router, _, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerSearchCandidates(t *testing.T) {
	router, _, candID := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/candidates?sort=relevance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp apisearch.SearchPage
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, candID.String(), resp.Items[0].UserId.String())
	assert.Equal(t, float32(94.0), resp.Items[0].Score)
	assert.Contains(t, resp.Items[0].Reasons, "top_test_performer")
}

func TestHandlerExplainCandidate(t *testing.T) {
	router, _, candID := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/candidates/"+candID.String()+"/explain", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp apisearch.ExplainResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.FinalScore)
	assert.True(t, *resp.FinalScore > 80.0)
	require.NotNil(t, resp.Factors)
	assert.Len(t, *resp.Factors, 4)
}

func TestHandlerCategoryRanking(t *testing.T) {
	router, repo, _ := setupTestRouter()
	catID := repo.docs[0].CategoryID

	req, _ := http.NewRequest(http.MethodGet, "/categories/"+catID.String()+"/ranking?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []apisearch.SearchHit
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, float32(94.0), resp[0].Score)
}
