package ranking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightsNormalization(t *testing.T) {
	w := Weights{
		WeightTest:     35,
		WeightFSP:      30,
		WeightStack:    25,
		WeightActivity: 10,
	}.Normalize()

	assert.InDelta(t, 0.35, w.WeightTest, 0.001)
	assert.InDelta(t, 0.30, w.WeightFSP, 0.001)
	assert.InDelta(t, 0.25, w.WeightStack, 0.001)
	assert.InDelta(t, 0.10, w.WeightActivity, 0.001)
}

func TestWeightsAdaptForQuery(t *testing.T) {
	w := DefaultWeights()

	// When query has no target stack
	adapted := w.AdaptForQuery(false, true)
	assert.Equal(t, 0.0, adapted.WeightStack)
	sum := adapted.WeightTest + adapted.WeightFSP + adapted.WeightActivity
	assert.InDelta(t, 1.0, sum, 0.001)
	assert.True(t, adapted.WeightTest > 0.35)
	assert.True(t, adapted.WeightFSP > 0.30)
}

func TestCalculateTopCandidate(t *testing.T) {
	goID := uuid.New()
	pgID := uuid.New()
	redisID := uuid.New()

	bestPlace := 2
	now := time.Now().Add(-2 * 24 * time.Hour) // 2 days ago

	cand := CandidateFeatures{
		UserID:               uuid.New(),
		DisplayName:          "Александр Смирнов",
		SpecializationName:   "Backend Go",
		GradeRank:            3,
		GradeName:            "Senior",
		TestScore:            96.5,
		TestPercentile:       98.0,
		HasFSP:               true,
		FSPScore:             88.0,
		FSPRating:            2480,
		SportsRank:           "Мастер спорта",
		FSPAchievementsCount: 5,
		FSPBestPlace:         &bestPlace,
		FSPWeightSum:         18,
		FSPHighlights:        []string{"серебряный призёр Кубка ФСП 2025"},
		CandidateStack:       []uuid.UUID{goID, pgID, redisID},
		CandidateStackNames:  []string{"Go", "PostgreSQL", "Redis"},
		PeriodicTasksSolved:  5,
		LastActiveAt:         &now,
	}

	params := RankingParams{
		TargetStack:      []uuid.UUID{goID, pgID, redisID},
		TargetStackNames: []string{"Go", "PostgreSQL", "Redis"},
	}

	result := Calculate(cand, params)

	// S_test = 96.5 -> 96.5 * 0.35 = 33.78
	// S_fsp = 88.0 -> 88.0 * 0.30 = 26.40
	// S_stack = 100.0 -> 100.0 * 0.25 = 25.00
	// S_act = (5 * 12) + 40 = 100.0 -> 100.0 * 0.10 = 10.00
	// Total expected approx: 33.78 + 26.40 + 25.00 + 10.00 = 95.18 -> 95.2

	require.NotNil(t, result)
	assert.InDelta(t, 95.2, result.FinalScore, 0.5)
	assert.Equal(t, float32(96.5), result.TestScore)
	assert.Equal(t, float32(88.0), result.FSPScore)
	assert.Equal(t, float32(100.0), result.StackScore)
	assert.Equal(t, float32(100.0), result.ActivityScore)

	// Check reasons
	assert.Contains(t, result.Reasons, "top_test_performer")
	assert.Contains(t, result.Reasons, "fsp_medalist")
	assert.Contains(t, result.Reasons, "fsp_master")
	assert.Contains(t, result.Reasons, "exact_stack_match")
	assert.Contains(t, result.Reasons, "active_solver")

	// Check human-readable explanation
	assert.Contains(t, result.Explanation, "Топ-5%")
	assert.Contains(t, result.Explanation, "серебряный призёр")
	assert.Contains(t, result.Explanation, "полное совпадение по стеку")
}

func TestCalculateCandidateWithoutFSP(t *testing.T) {
	pyID := uuid.New()
	pgID := uuid.New()
	dockerID := uuid.New()

	cand := CandidateFeatures{
		UserID:              uuid.New(),
		DisplayName:         "Иван Петров",
		SpecializationName:  "Backend Python",
		GradeRank:           2,
		TestScore:           92.0,
		TestPercentile:      92.0,
		HasFSP:              false,
		FSPScore:            0.0,
		CandidateStack:      []uuid.UUID{pyID, pgID, dockerID},
		CandidateStackNames: []string{"Python", "PostgreSQL", "Docker"},
		PeriodicTasksSolved: 3,
	}

	params := RankingParams{
		TargetStack:      []uuid.UUID{pyID, pgID, dockerID},
		TargetStackNames: []string{"Python", "PostgreSQL", "Docker"},
	}

	result := Calculate(cand, params)

	require.NotNil(t, result)
	assert.Equal(t, float32(0.0), result.FSPScore)
	assert.True(t, result.FinalScore > 50.0) // Still high from test, stack, and activity

	// Check reasons
	assert.Contains(t, result.Reasons, "no_fsp_candidate")
	assert.Contains(t, result.Reasons, "exact_stack_match")
	assert.NotContains(t, result.Reasons, "fsp_verified")

	// Check factors
	var fspFactor *ScoreFactor
	for _, f := range result.Factors {
		if f.Name == "fsp_achievements" {
			fspFactor = &f
			break
		}
	}
	require.NotNil(t, fspFactor)
	assert.Equal(t, float32(0.0), fspFactor.RawScore)
	assert.Contains(t, fspFactor.Detail, "История участия в соревнованиях ФСП не привязана")
}

func TestCalculatePartialStackMatch(t *testing.T) {
	skill1 := uuid.New()
	skill2 := uuid.New()
	skill3 := uuid.New()

	cand := CandidateFeatures{
		UserID:         uuid.New(),
		TestScore:      75.0,
		HasFSP:         false,
		CandidateStack: []uuid.UUID{skill1, skill2}, // matches 2 out of 3
	}

	params := RankingParams{
		TargetStack: []uuid.UUID{skill1, skill2, skill3},
	}

	result := Calculate(cand, params)
	assert.InDelta(t, 66.7, result.StackScore, 0.5)
}
