package ranking

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// CalculateComputes calculates the four individual scores:
// S_test, S_fsp, S_stack, S_act and the final weighted Score.
func Calculate(c CandidateFeatures, p RankingParams) RankingScore {
	// 1. Determine active weights
	weights := DefaultWeights()
	if p.CustomWeights != nil {
		weights = *p.CustomWeights
	}

	hasTargetStack := len(p.TargetStack) > 0
	if p.AdaptiveWeights {
		weights = weights.AdaptForQuery(hasTargetStack, c.HasFSP)
	} else {
		weights = weights.Normalize()
	}

	// 2. Component Scores S_i \in [0, 100]
	sTest := calculateTestScore(c)
	sFSP := calculateFSPScore(c)
	sStack, matchedStackCount, totalTargetCount := calculateStackScore(c, p.TargetStack)
	sAct := calculateActivityScore(c)

	// 3. Weighted contributions
	contribTest := float32(math.Round(weights.WeightTest*sTest*100) / 100)
	contribFSP := float32(math.Round(weights.WeightFSP*sFSP*100) / 100)
	contribStack := float32(math.Round(weights.WeightStack*sStack*100) / 100)
	contribAct := float32(math.Round(weights.WeightActivity*sAct*100) / 100)

	finalScore := contribTest + contribFSP + contribStack + contribAct
	if finalScore > 100.0 {
		finalScore = 100.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}
	finalScore = float32(math.Round(float64(finalScore)*10) / 10)

	// 4. Generate Explainability details and tags
	engine := NewExplainabilityEngine()
	factors, reasons, explanation := engine.Generate(c, p, ExplainInputs{
		Weights:            weights,
		STest:              float32(sTest),
		SFSP:               float32(sFSP),
		SStack:             float32(sStack),
		SAct:               float32(sAct),
		ContribTest:        contribTest,
		ContribFSP:         contribFSP,
		ContribStack:       contribStack,
		ContribAct:         contribAct,
		MatchedStackCount:  matchedStackCount,
		TotalTargetCount:   totalTargetCount,
		FinalScore:         finalScore,
	})

	return RankingScore{
		FinalScore:    finalScore,
		TestScore:     float32(sTest),
		FSPScore:      float32(sFSP),
		StackScore:    float32(sStack),
		ActivityScore: float32(sAct),
		Factors:       factors,
		Reasons:       reasons,
		Explanation:   explanation,
	}
}

// calculateTestScore computes S_test \in [0, 100]
func calculateTestScore(c CandidateFeatures) float64 {
	score := c.TestScore
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return math.Round(score*10) / 10
}

// calculateFSPScore computes S_fsp \in [0, 100]
func calculateFSPScore(c CandidateFeatures) float64 {
	if !c.HasFSP {
		return 0.0
	}
	score := c.FSPScore
	if score <= 0 && (c.FSPWeightSum > 0 || c.SportsRank != "") {
		// Fallback baseline if FSPScore wasn't pre-computed
		score = float64(c.FSPWeightSum) * 2.0
		if score > 50 {
			score = 50
		}
	}
	if score > 100 {
		score = 100
	}
	return math.Round(score*10) / 10
}

// calculateStackScore computes S_stack \in [0, 100] based on target stack overlap
func calculateStackScore(c CandidateFeatures, targetStack []uuid.UUID) (score float64, matchedCount int, totalCount int) {
	if len(targetStack) == 0 {
		// If employer did not specify required stack: evaluate candidate's tech breadth
		skillCount := len(c.CandidateStack)
		switch {
		case skillCount >= 4:
			return 100.0, skillCount, skillCount
		case skillCount == 3:
			return 85.0, skillCount, 4
		case skillCount == 2:
			return 70.0, skillCount, 4
		case skillCount == 1:
			return 50.0, skillCount, 4
		default:
			return 30.0, 0, 4
		}
	}

	candSet := make(map[uuid.UUID]struct{}, len(c.CandidateStack))
	for _, id := range c.CandidateStack {
		candSet[id] = struct{}{}
	}

	matches := 0
	for _, reqID := range targetStack {
		if _, ok := candSet[reqID]; ok {
			matches++
		}
	}

	ratio := float64(matches) / float64(len(targetStack))
	score = ratio * 100.0
	return math.Round(score*10) / 10, matches, len(targetStack)
}

// calculateActivityScore computes S_act \in [0, 100]
func calculateActivityScore(c CandidateFeatures) float64 {
	var total float64

	// 1. Periodic micro-tasks solved on platform (up to 60 pts)
	// 12 points per task, max 5 tasks = 60 pts
	tasksScore := float64(c.PeriodicTasksSolved) * 12.0
	if tasksScore > 60.0 {
		tasksScore = 60.0
	}
	total += tasksScore

	// 2. Activity recency / freshness (up to 40 pts)
	if c.LastActiveAt != nil {
		diff := time.Since(*c.LastActiveAt)
		switch {
		case diff <= 7*24*time.Hour:
			total += 40.0
		case diff <= 30*24*time.Hour:
			total += 30.0
		case diff <= 90*24*time.Hour:
			total += 20.0
		default:
			total += 10.0
		}
	} else {
		// Default base recency for active registered profile
		total += 20.0
	}

	if total > 100.0 {
		total = 100.0
	}
	return math.Round(total*10) / 10
}
