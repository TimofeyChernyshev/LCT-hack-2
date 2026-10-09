package fsp

import (
	"fmt"
	"math"
	"strings"
)

const (
	// NoFSPExplanation is the transparent explanation string for candidates without FSP history
	NoFSPExplanation = "История участия в соревнованиях ФСП не привязана. Кандидат оценивается по результатам тестов и стеку компетенций."
)

// CalculateScore computes S_fsp \in [0, 100], metrics, and human-readable explanation
func CalculateScore(member *Member) ScoreResult {
	if member == nil || strings.TrimSpace(member.FSPID) == "" {
		return ScoreResult{
			Score:             0.0,
			WeightSum:         0,
			AchievementsCount: 0,
			BestPlace:         nil,
			Explanation:       NoFSPExplanation,
		}
	}

	var totalScore float64
	var weightSum int
	var bestPlace *int

	// 1. Sports Rank score (max 40 pts)
	switch member.SportsRank {
	case RankZMS, RankMSMK:
		totalScore += 40.0
	case RankMS:
		totalScore += 35.0
	case RankKMS:
		totalScore += 28.0
	case Rank1:
		totalScore += 20.0
	case Rank2:
		totalScore += 14.0
	case Rank3:
		totalScore += 8.0
	default:
		if member.SportsRank != "" && member.SportsRank != RankUnranked {
			totalScore += 5.0
		} else {
			totalScore += 2.0
		}
	}

	// 2. Rating component (max 20 pts)
	// Rating scale approx 1200..2800
	if member.Rating > 1200 {
		normalizedRating := float64(member.Rating-1200) / 1600.0 * 20.0
		if normalizedRating > 20.0 {
			normalizedRating = 20.0
		}
		totalScore += normalizedRating
	}

	// 3. Achievements & Tournaments (max 40 pts)
	var achievementBonus float64
	var topHighlights []string

	for _, ach := range member.Achievements {
		weightSum += ach.Weight

		// Track best place
		if ach.Place != nil && *ach.Place > 0 {
			if bestPlace == nil || *ach.Place < *bestPlace {
				bp := *ach.Place
				bestPlace = &bp
			}

			// Bonus points for podium places
			if *ach.Place == 1 {
				if ach.Category == CategoryChampionship || ach.Category == CategoryCup {
					achievementBonus += 10.0
				} else if ach.Category == CategoryHackathon {
					achievementBonus += 8.0
				} else {
					achievementBonus += 5.0
				}
			} else if *ach.Place <= 3 {
				achievementBonus += 5.0
			} else if *ach.Place <= 10 {
				achievementBonus += 2.0
			}
		}

		// Weight contribution
		achievementBonus += float64(ach.Weight) * 1.5

		// Collect top highlights for explanation
		if len(topHighlights) < 3 {
			highlight := formatAchievementHighlight(ach)
			if highlight != "" {
				topHighlights = append(topHighlights, highlight)
			}
		}
	}

	if achievementBonus > 40.0 {
		achievementBonus = 40.0
	}
	totalScore += achievementBonus

	// Cap final score at 100.0
	if totalScore > 100.0 {
		totalScore = 100.0
	}
	totalScore = math.Round(totalScore*10) / 10

	// 4. Generate Explainability text
	explanation := buildExplanation(member, topHighlights, weightSum, bestPlace)

	return ScoreResult{
		Score:             totalScore,
		WeightSum:         weightSum,
		AchievementsCount: len(member.Achievements),
		BestPlace:         bestPlace,
		Explanation:       explanation,
	}
}

func formatAchievementHighlight(ach Achievement) string {
	if ach.Place != nil && *ach.Place > 0 {
		switch *ach.Place {
		case 1:
			return fmt.Sprintf("1 место: %s", ach.EventName)
		case 2:
			return fmt.Sprintf("серебряный призёр: %s", ach.EventName)
		case 3:
			return fmt.Sprintf("бронзовый призёр: %s", ach.EventName)
		default:
			return fmt.Sprintf("%d место: %s", *ach.Place, ach.EventName)
		}
	}
	if ach.Category == CategoryHackathon {
		return fmt.Sprintf("участник хакатона %s", ach.EventName)
	}
	return ach.EventName
}

func buildExplanation(member *Member, highlights []string, weightSum int, bestPlace *int) string {
	var parts []string

	// Rank and Rating
	rankStr := string(member.SportsRank)
	if rankStr == "" {
		rankStr = "Участник ФСП"
	}

	if member.Rating > 0 {
		parts = append(parts, fmt.Sprintf("%s (рейтинг ФСП: %d)", rankStr, member.Rating))
	} else {
		parts = append(parts, rankStr)
	}

	// Highlights
	if len(highlights) > 0 {
		parts = append(parts, strings.Join(highlights, ", "))
	}

	// Weight summary
	if weightSum > 0 {
		parts = append(parts, fmt.Sprintf("суммарный вес спортивных достижений: %d", weightSum))
	} else if len(member.Achievements) == 0 {
		parts = append(parts, "официальный статус в реестре подтвержден")
	}

	return strings.Join(parts, ". ") + "."
}
