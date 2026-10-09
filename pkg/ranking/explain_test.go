package ranking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplainabilityGenerationFullProfile(t *testing.T) {
	firstPlace := 1
	recent := time.Now().Add(-24 * time.Hour)

	cand := CandidateFeatures{
		UserID:               uuid.New(),
		DisplayName:          "Дмитрий Васильев",
		SpecializationName:   "Backend Python",
		GradeName:            "Senior",
		TestScore:            98.0,
		TestPercentile:       99.0,
		HasFSP:               true,
		FSPScore:             95.0,
		FSPRating:            2600,
		SportsRank:           "ЗМС",
		FSPAchievementsCount: 8,
		FSPBestPlace:         &firstPlace,
		FSPWeightSum:         32,
		FSPHighlights:        []string{"победитель Чемпионата России ФСП 2024"},
		CandidateStackNames:  []string{"Python", "FastAPI", "PostgreSQL", "Redis"},
		PeriodicTasksSolved:  6,
		LastActiveAt:         &recent,
	}

	params := RankingParams{
		TargetStackNames: []string{"Python", "FastAPI", "PostgreSQL", "Redis"},
	}

	engine := NewExplainabilityEngine()
	factors, reasons, exp := engine.Generate(cand, params, ExplainInputs{
		Weights:            DefaultWeights(),
		STest:              98.0,
		SFSP:               95.0,
		SStack:             100.0,
		SAct:               100.0,
		ContribTest:        34.3,
		ContribFSP:         28.5,
		ContribStack:       25.0,
		ContribAct:         10.0,
		MatchedStackCount:  4,
		TotalTargetCount:   4,
		FinalScore:         97.8,
	})

	require.Len(t, factors, 4)
	assert.Contains(t, reasons, "top_test_performer")
	assert.Contains(t, reasons, "fsp_champion")
	assert.Contains(t, reasons, "fsp_medalist")
	assert.Contains(t, reasons, "fsp_master")
	assert.Contains(t, reasons, "exact_stack_match")
	assert.Contains(t, reasons, "active_solver")
	assert.Contains(t, reasons, "recently_active")

	assert.Contains(t, exp, "Топ-5%")
	assert.Contains(t, exp, "победитель")
	assert.Contains(t, exp, "полное совпадение по стеку")
}

func TestExplainabilityGenerationNoFSP(t *testing.T) {
	cand := CandidateFeatures{
		UserID:              uuid.New(),
		DisplayName:         "Елена Смирнова",
		SpecializationName:  "Frontend React",
		TestScore:           85.0,
		TestPercentile:      88.0,
		HasFSP:              false,
		CandidateStackNames: []string{"React", "TypeScript"},
	}

	params := RankingParams{
		TargetStackNames: []string{"React", "TypeScript"},
	}

	engine := NewExplainabilityEngine()
	factors, reasons, exp := engine.Generate(cand, params, ExplainInputs{
		Weights:           DefaultWeights(),
		STest:             85.0,
		SFSP:              0.0,
		SStack:            100.0,
		SAct:              20.0,
		ContribTest:       29.75,
		ContribFSP:        0.0,
		ContribStack:      25.0,
		ContribAct:        2.0,
		MatchedStackCount: 2,
		TotalTargetCount:  2,
		FinalScore:        56.75,
	})

	require.Len(t, factors, 4)
	assert.Contains(t, reasons, "no_fsp_candidate")
	assert.NotContains(t, reasons, "fsp_verified")

	// FSP factor explanation must transparently explain the lack of FSP history
	var fspFactor *ScoreFactor
	for _, f := range factors {
		if f.Name == "fsp_achievements" {
			fspFactor = &f
			break
		}
	}
	require.NotNil(t, fspFactor)
	assert.Contains(t, fspFactor.Detail, "История участия в соревнованиях ФСП не привязана")
	assert.NotEmpty(t, exp)
}
