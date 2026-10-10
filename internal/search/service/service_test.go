package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/config"
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
	var res []domain.CandidateSearchDoc
	for _, d := range m.docs {
		if d.CategoryID == categoryID {
			res = append(res, d)
		}
	}
	return res, nil
}

func (m *mockRepo) CalculateTestPercentile(ctx context.Context, categoryID uuid.UUID, testScore float64) (float64, error) {
	if testScore >= 90 {
		return 95.0, nil
	}
	return 75.0, nil
}

func (m *mockRepo) Upsert(ctx context.Context, doc *domain.CandidateSearchDoc) error {
	m.docs = append(m.docs, *doc)
	return nil
}

func (m *mockRepo) Count(ctx context.Context) (int, error) {
	return len(m.docs), nil
}

func TestSearchServiceSearchCandidates(t *testing.T) {
	catID := uuid.New()
	specID := uuid.New()
	gradeID := uuid.New()
	cand1ID := uuid.New()
	cand2ID := uuid.New()

	goTech := uuid.New()
	pgTech := uuid.New()

	testScore1 := 95.0
	testScore2 := 70.0
	place1 := 1

	now := time.Now()

	repo := &mockRepo{
		docs: []domain.CandidateSearchDoc{
			{
				UserID:               cand1ID,
				DisplayName:          "Алексей Иванов",
				CategoryID:           catID,
				SpecializationID:     specID,
				SpecializationName:   "Backend Go",
				GradeID:              gradeID,
				GradeName:            "Senior",
				TestScore:            &testScore1,
				HasFSP:               true,
				SportsRank:           "Мастер спорта",
				FSPScore:             85.0,
				FSPBestPlace:         &place1,
				FSPWeightSum:         20,
				Stack:                []uuid.UUID{goTech, pgTech},
				PeriodicTasksSolved:  4,
				CalculatedScore:      92.5,
				Explanation:          "Топ-5% по тесту Go, 1 место ФСП",
				Reasons:              []string{"top_test_performer", "fsp_champion"},
				LastActiveAt:         now,
			},
			{
				UserID:              cand2ID,
				DisplayName:         "Борис Новиков",
				CategoryID:          catID,
				SpecializationID:    specID,
				SpecializationName:  "Backend Go",
				GradeID:             gradeID,
				GradeName:           "Middle",
				TestScore:           &testScore2,
				HasFSP:              false,
				Stack:               []uuid.UUID{goTech},
				PeriodicTasksSolved: 1,
				CalculatedScore:     65.0,
				Explanation:         "Базовый уровень тестирования",
				Reasons:             []string{"no_fsp_candidate"},
				LastActiveAt:        now,
			},
		},
	}

	cfg := &config.Config{DefaultLimit: 20, MaxLimit: 100}
	svc := NewSearchService(repo, cfg)

	// Search without target stack filter
	page, err := svc.SearchCandidates(context.Background(), domain.SearchFilter{
		CategoryID: &catID,
	})

	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Total)
	assert.Len(t, page.Items, 2)

	assert.Equal(t, cand1ID, page.Items[0].UserID)
	assert.Equal(t, float32(92.5), page.Items[0].Score)
	assert.Contains(t, page.Items[0].Reasons, "top_test_performer")

	// Search with target stack filter -> triggers dynamic re-ranking
	pageTarget, err := svc.SearchCandidates(context.Background(), domain.SearchFilter{
		CategoryID: &catID,
		Stack:      []uuid.UUID{goTech, pgTech},
	})
	require.NoError(t, err)
	assert.Len(t, pageTarget.Items, 2)
	assert.True(t, pageTarget.Items[0].Score > pageTarget.Items[1].Score)
	assert.Contains(t, pageTarget.Items[0].Reasons, "exact_stack_match")
}

func TestSearchServiceExplainCandidate(t *testing.T) {
	catID := uuid.New()
	userID := uuid.New()
	testScore := 96.0
	place := 2
	goTech := uuid.New()

	repo := &mockRepo{
		docs: []domain.CandidateSearchDoc{
			{
				UserID:               userID,
				DisplayName:          "Дмитрий Соколов",
				CategoryID:           catID,
				SpecializationName:   "Backend Go",
				TestScore:            &testScore,
				HasFSP:               true,
				SportsRank:           "КМС",
				FSPScore:             80.0,
				FSPBestPlace:         &place,
				FSPWeightSum:         15,
				Stack:                []uuid.UUID{goTech},
				PeriodicTasksSolved:  3,
				LastActiveAt:         time.Now(),
			},
		},
	}

	cfg := &config.Config{DefaultLimit: 20, MaxLimit: 100}
	svc := NewSearchService(repo, cfg)

	explain, err := svc.ExplainCandidate(context.Background(), userID, []uuid.UUID{goTech})
	require.NoError(t, err)
	require.NotNil(t, explain)
	assert.Equal(t, userID, explain.UserID)
	assert.True(t, explain.FinalScore > 75.0)
	assert.Len(t, explain.Factors, 4)
	assert.Contains(t, explain.Reasons, "top_test_performer")
	assert.Contains(t, explain.Reasons, "fsp_medalist")
	assert.NotEmpty(t, explain.Summary)
}

func TestSearchServiceCategoryRanking(t *testing.T) {
	catID := uuid.New()
	userID := uuid.New()

	repo := &mockRepo{
		docs: []domain.CandidateSearchDoc{
			{
				UserID:          userID,
				CategoryID:      catID,
				CalculatedScore: 89.0,
				Explanation:     "Высокий балл",
				Reasons:         []string{"top_test_performer"},
			},
		},
	}

	cfg := &config.Config{DefaultLimit: 20, MaxLimit: 100}
	svc := NewSearchService(repo, cfg)

	ranking, err := svc.GetCategoryRanking(context.Background(), catID, 10)
	require.NoError(t, err)
	require.Len(t, ranking, 1)
	assert.Equal(t, userID, ranking[0].UserID)
	assert.Equal(t, float32(89.0), ranking[0].Score)
}
