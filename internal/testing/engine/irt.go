package engine

import (
	"math"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
)

type ItemEvaluation struct {
	Discrimination float64 // a
	Difficulty     float64 // b
	Score          float64 // u: 0.0 .. 1.0
}

type IRTScorer struct {
	PassThresholdRatio    float64
	UpgradeThresholdRatio float64
}

func NewIRTScorer(passThreshold, upgradeThreshold float64) *IRTScorer {
	if passThreshold <= 0 {
		passThreshold = 0.7
	}
	if upgradeThreshold <= 0 {
		upgradeThreshold = 0.9
	}
	return &IRTScorer{
		PassThresholdRatio:    passThreshold,
		UpgradeThresholdRatio: upgradeThreshold,
	}
}

// MapDifficultyToIRT converts discrete difficulty (1..5) to IRT b scale (-1.5 .. +1.5) if not set
func MapDifficultyToIRT(difficulty int, customIRT float64) float64 {
	if customIRT != 0 {
		return customIRT
	}
	switch difficulty {
	case 1:
		return -1.5
	case 2:
		return -0.75
	case 3:
		return 0.0
	case 4:
		return 0.75
	case 5:
		return 1.5
	default:
		return 0.0
	}
}

// EstimateAbility calculates latent ability theta (-3.0 to +3.0) using 2PL IRT model
func (s *IRTScorer) EstimateAbility(items []ItemEvaluation) float64 {
	if len(items) == 0 {
		return 0.0
	}

	// Newton-Raphson to solve sum(a_i * (u_i - P_i(theta))) = 0
	theta := 0.0
	const maxIter = 25
	const tolerance = 1e-4

	for iter := 0; iter < maxIter; iter++ {
		scoreFunc := 0.0
		infoFunc := 0.0

		for _, item := range items {
			a := item.Discrimination
			if a <= 0 {
				a = 1.0
			}
			b := item.Difficulty
			p := 1.0 / (1.0 + math.Exp(-a*(theta-b)))
			q := 1.0 - p

			scoreFunc += a * (item.Score - p)
			infoFunc += a * a * p * q
		}

		// Regularize info to avoid division by near-zero
		infoFunc += 0.1

		delta := scoreFunc / infoFunc
		theta += delta

		if math.Abs(delta) < tolerance {
			break
		}

		// Constrain within bounds
		if theta > 3.0 {
			theta = 3.0
			break
		} else if theta < -3.0 {
			theta = -3.0
			break
		}
	}

	return math.Round(theta*1000) / 1000
}

// ComputeSessionResult computes overall score (0..100) and grade decision
func (s *IRTScorer) ComputeSessionResult(items []ItemEvaluation) (score float64, theta float64, decision string) {
	if len(items) == 0 {
		return 0.0, 0.0, string(domain.SessionStatusCancelled)
	}

	totalScore := 0.0
	for _, it := range items {
		totalScore += it.Score
	}

	ratio := totalScore / float64(len(items))
	score = math.Round(ratio*10000) / 100 // 0.00 .. 100.00%
	theta = s.EstimateAbility(items)

	switch {
	case ratio >= s.UpgradeThresholdRatio:
		decision = "upgrade_offered"
	case ratio >= s.PassThresholdRatio:
		decision = "confirmed"
	default:
		decision = "downgrade_offered"
	}

	return score, theta, decision
}
