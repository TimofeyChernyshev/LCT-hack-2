package ranking

import (
	"time"

	"github.com/google/uuid"
)

// Weights defines importance of each component in the ranking formula:
// Score = w_test * S_test + w_fsp * S_fsp + w_stack * S_stack + w_act * S_act
type Weights struct {
	WeightTest     float64 `json:"weightTest"`     // w_test, default 0.35
	WeightFSP      float64 `json:"weightFsp"`      // w_fsp, default 0.30
	WeightStack    float64 `json:"weightStack"`    // w_stack, default 0.25
	WeightActivity float64 `json:"weightActivity"` // w_act, default 0.10
}

// DefaultWeights returns the standard weights aligned with FSP platform requirements
func DefaultWeights() Weights {
	return Weights{
		WeightTest:     0.35,
		WeightFSP:      0.30,
		WeightStack:    0.25,
		WeightActivity: 0.10,
	}
}

// Normalize ensures the sum of weights equals 1.0
func (w Weights) Normalize() Weights {
	sum := w.WeightTest + w.WeightFSP + w.WeightStack + w.WeightActivity
	if sum <= 0 {
		return DefaultWeights()
	}
	return Weights{
		WeightTest:     w.WeightTest / sum,
		WeightFSP:      w.WeightFSP / sum,
		WeightStack:    w.WeightStack / sum,
		WeightActivity: w.WeightActivity / sum,
	}
}

// AdaptForQuery dynamically rebalances weights if specific query requirements are absent.
// e.g. If employer specified no stack filter, stack weight is redistributed across test, FSP and activity.
func (w Weights) AdaptForQuery(hasTargetStack bool, hasFSPHistory bool) Weights {
	normalized := w.Normalize()

	// If no target stack was requested by employer
	if !hasTargetStack {
		remWeight := normalized.WeightStack
		totalOther := normalized.WeightTest + normalized.WeightFSP + normalized.WeightActivity
		if totalOther > 0 {
			normalized.WeightTest += remWeight * (normalized.WeightTest / totalOther)
			normalized.WeightFSP += remWeight * (normalized.WeightFSP / totalOther)
			normalized.WeightActivity += remWeight * (normalized.WeightActivity / totalOther)
			normalized.WeightStack = 0.0
		}
	}

	return normalized.Normalize()
}

// CandidateFeatures contains all objective signals extracted for candidate ranking
type CandidateFeatures struct {
	UserID               uuid.UUID    `json:"userId"`
	DisplayName          string       `json:"displayName"`
	CategoryID           uuid.UUID    `json:"categoryId"`
	CategoryName         string       `json:"categoryName,omitempty"`
	SpecializationID     uuid.UUID    `json:"specializationId"`
	SpecializationName   string       `json:"specializationName,omitempty"`
	GradeID              uuid.UUID    `json:"gradeId"`
	GradeName            string       `json:"gradeName,omitempty"`
	GradeRank            int          `json:"gradeRank"` // 1=Junior, 2=Middle, 3=Senior, 4=Lead
	TestScore            float64      `json:"testScore"` // S_test \in [0, 100]
	TestPercentile       float64      `json:"testPercentile"` // e.g. 95.0 = top 5%
	HasFSP               bool         `json:"hasFsp"`
	FSPScore             float64      `json:"fspScore"` // S_fsp \in [0, 100]
	FSPRating            int          `json:"fspRating"`
	SportsRank           string       `json:"sportsRank,omitempty"`
	FSPAchievementsCount int          `json:"fspAchievementsCount"`
	FSPBestPlace         *int         `json:"fspBestPlace,omitempty"`
	FSPWeightSum         int          `json:"fspWeightSum"`
	FSPHighlights        []string     `json:"fspHighlights,omitempty"`
	CandidateStack       []uuid.UUID  `json:"candidateStack"`
	CandidateStackNames  []string     `json:"candidateStackNames,omitempty"`
	YearsExperience      *float64     `json:"yearsExperience,omitempty"`
	PeriodicTasksSolved  int          `json:"periodicTasksSolved"`
	LastActiveAt         *time.Time   `json:"lastActiveAt,omitempty"`
}

// RankingParams holds criteria from the employer's search query
type RankingParams struct {
	TargetStack          []uuid.UUID `json:"targetStack,omitempty"`
	TargetStackNames     []string    `json:"targetStackNames,omitempty"`
	CustomWeights        *Weights    `json:"customWeights,omitempty"`
	AdaptiveWeights      bool        `json:"adaptiveWeights"`
}

// ScoreFactor represents one factor contribution in the Explainable AI (XAI) output
type ScoreFactor struct {
	Name         string  `json:"name"`         // e.g. "test_score", "fsp_achievements", "stack_match", "platform_activity"
	Weight       float32 `json:"weight"`       // Factor weight w_i
	RawScore     float32 `json:"rawScore"`     // S_i \in [0, 100]
	Contribution float32 `json:"contribution"` // w_i * S_i
	Detail       string  `json:"detail"`       // Clear Russian explanation of this factor
}

// RankingScore is the full evaluated ranking result for a candidate
type RankingScore struct {
	FinalScore    float32       `json:"finalScore"`    // Overall weighted score in [0, 100]
	TestScore     float32       `json:"testScore"`     // S_test
	FSPScore      float32       `json:"fspScore"`      // S_fsp
	StackScore    float32       `json:"stackScore"`    // S_stack
	ActivityScore float32       `json:"activityScore"` // S_act
	Factors       []ScoreFactor `json:"factors"`       // Detailed explainability breakdown
	Reasons       []string      `json:"reasons"`       // Badge tags (e.g. "top_test_performer", "fsp_medalist")
	Explanation   string        `json:"explanation"`   // Synthesized human-readable rationale ("Почему кандидат в топе")
}
