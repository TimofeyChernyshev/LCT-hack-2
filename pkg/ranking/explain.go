package ranking

import (
	"fmt"
	"strings"
	"time"
)

// ExplainInputs contains calculated scores and contributions for explanation synthesis
type ExplainInputs struct {
	Weights           Weights
	STest             float32
	SFSP              float32
	SStack            float32
	SAct              float32
	ContribTest       float32
	ContribFSP        float32
	ContribStack      float32
	ContribAct        float32
	MatchedStackCount int
	TotalTargetCount  int
	FinalScore        float32
}

// ExplainabilityEngine produces structured XAI factors and natural Russian justification
type ExplainabilityEngine struct{}

func NewExplainabilityEngine() *ExplainabilityEngine {
	return &ExplainabilityEngine{}
}

// Generate builds the factors, badge reason tags, and synthesized explanation text
func (e *ExplainabilityEngine) Generate(c CandidateFeatures, p RankingParams, in ExplainInputs) ([]ScoreFactor, []string, string) {
	factors := make([]ScoreFactor, 0, 4)
	reasons := make([]string, 0, 8)
	var topHighlights []string

	// ----------------------------------------------------
	// 1. Factor: Test Score
	// ----------------------------------------------------
	testDetail, testHighlight, testReasons := e.buildTestFactor(c, in)
	factors = append(factors, ScoreFactor{
		Name:         "test_score",
		Weight:       float32(in.Weights.WeightTest),
		RawScore:     in.STest,
		Contribution: in.ContribTest,
		Detail:       testDetail,
	})
	if testHighlight != "" {
		topHighlights = append(topHighlights, testHighlight)
	}
	reasons = append(reasons, testReasons...)

	// ----------------------------------------------------
	// 2. Factor: FSP Sports Track
	// ----------------------------------------------------
	fspDetail, fspHighlight, fspReasons := e.buildFSPFactor(c, in)
	factors = append(factors, ScoreFactor{
		Name:         "fsp_achievements",
		Weight:       float32(in.Weights.WeightFSP),
		RawScore:     in.SFSP,
		Contribution: in.ContribFSP,
		Detail:       fspDetail,
	})
	if fspHighlight != "" {
		topHighlights = append(topHighlights, fspHighlight)
	}
	reasons = append(reasons, fspReasons...)

	// ----------------------------------------------------
	// 3. Factor: Stack Match
	// ----------------------------------------------------
	stackDetail, stackHighlight, stackReasons := e.buildStackFactor(c, p, in)
	factors = append(factors, ScoreFactor{
		Name:         "stack_match",
		Weight:       float32(in.Weights.WeightStack),
		RawScore:     in.SStack,
		Contribution: in.ContribStack,
		Detail:       stackDetail,
	})
	if stackHighlight != "" {
		topHighlights = append(topHighlights, stackHighlight)
	}
	reasons = append(reasons, stackReasons...)

	// ----------------------------------------------------
	// 4. Factor: Platform Activity
	// ----------------------------------------------------
	actDetail, actHighlight, actReasons := e.buildActivityFactor(c, in)
	factors = append(factors, ScoreFactor{
		Name:         "platform_activity",
		Weight:       float32(in.Weights.WeightActivity),
		RawScore:     in.SAct,
		Contribution: in.ContribAct,
		Detail:       actDetail,
	})
	if actHighlight != "" {
		topHighlights = append(topHighlights, actHighlight)
	}
	reasons = append(reasons, actReasons...)

	// ----------------------------------------------------
	// Synthesized Explanation ("Почему кандидат в топе")
	// ----------------------------------------------------
	var explanation string
	if len(topHighlights) > 0 {
		explanation = strings.Join(topHighlights, ", ") + "."
	} else {
		explanation = fmt.Sprintf("Кандидат с подтвержденной квалификацией (итоговый скоринг: %.1f).", in.FinalScore)
	}

	return factors, reasons, explanation
}

func (e *ExplainabilityEngine) buildTestFactor(c CandidateFeatures, in ExplainInputs) (detail string, highlight string, reasons []string) {
	specName := c.SpecializationName
	if specName == "" {
		specName = "профильному направлению"
	}

	percentile := c.TestPercentile
	if percentile <= 0 && in.STest >= 90 {
		percentile = 95.0
	} else if percentile <= 0 && in.STest >= 80 {
		percentile = 85.0
	}

	switch {
	case percentile >= 95:
		detail = fmt.Sprintf("Топ-5%% по результатам теста на знание %s (балл %.1f%%)", specName, in.STest)
		highlight = fmt.Sprintf("Топ-5%% по результатам теста на знание %s (балл %.1f%%)", specName, in.STest)
		reasons = append(reasons, "top_test_performer")
	case percentile >= 90:
		detail = fmt.Sprintf("Топ-10%% по результатам теста на знание %s (балл %.1f%%)", specName, in.STest)
		highlight = fmt.Sprintf("Топ-10%% по результатам теста на знание %s (балл %.1f%%)", specName, in.STest)
		reasons = append(reasons, "top_test_performer")
	case in.STest >= 80:
		detail = fmt.Sprintf("Высокий результат теста на знание %s (балл %.1f%%)", specName, in.STest)
		highlight = fmt.Sprintf("высокий балл теста на знание %s (%.1f%%)", specName, in.STest)
		reasons = append(reasons, "high_test_score")
	case in.STest >= 60:
		detail = fmt.Sprintf("Успешно подтверждена квалификация по тесту %s (балл %.1f%%)", specName, in.STest)
		highlight = fmt.Sprintf("успешное прохождение теста %s (%.1f%%)", specName, in.STest)
		reasons = append(reasons, "test_passed")
	default:
		detail = fmt.Sprintf("Базовый уровень тестирования по %s (балл %.1f%%)", specName, in.STest)
	}

	return detail, highlight, reasons
}

func (e *ExplainabilityEngine) buildFSPFactor(c CandidateFeatures, in ExplainInputs) (detail string, highlight string, reasons []string) {
	if !c.HasFSP {
		detail = "История участия в соревнованиях ФСП не привязана. Кандидат оценивается по результатам тестов и стеку компетенций."
		reasons = append(reasons, "no_fsp_candidate")
		return detail, "", reasons
	}

	reasons = append(reasons, "fsp_verified")

	// Check podium and sports ranks
	if c.FSPBestPlace != nil && *c.FSPBestPlace > 0 {
		switch *c.FSPBestPlace {
		case 1:
			reasons = append(reasons, "fsp_champion", "fsp_medalist")
		case 2, 3:
			reasons = append(reasons, "fsp_medalist")
		}
	}

	if c.SportsRank != "" {
		switch c.SportsRank {
		case "ЗМС", "МСМК", "Мастер спорта":
			reasons = append(reasons, "fsp_master")
		case "КМС", "1-й спортивный разряд", "2-й спортивный разряд", "3-й спортивный разряд":
			reasons = append(reasons, "fsp_ranked")
		}
	}

	var parts []string
	if len(c.FSPHighlights) > 0 {
		parts = append(parts, c.FSPHighlights[0])
	} else if c.FSPBestPlace != nil && *c.FSPBestPlace > 0 {
		switch *c.FSPBestPlace {
		case 1:
			parts = append(parts, "победитель соревнований ФСП")
		case 2:
			parts = append(parts, "серебряный призёр соревнований ФСП")
		case 3:
			parts = append(parts, "бронзовый призёр соревнований ФСП")
		default:
			parts = append(parts, fmt.Sprintf("%d место на соревнованиях ФСП", *c.FSPBestPlace))
		}
	}

	if c.SportsRank != "" && c.SportsRank != "Без разряда" {
		parts = append(parts, c.SportsRank)
	}

	if c.FSPWeightSum > 0 {
		parts = append(parts, fmt.Sprintf("суммарный вес спортивных достижений: %d", c.FSPWeightSum))
	}

	if len(parts) == 0 {
		detail = "Подтвержденный статус участника в реестре ФСП"
		highlight = "верифицированный профиль ФСП"
	} else {
		detail = strings.Join(parts, ", ")
		highlight = detail
	}

	return detail, highlight, reasons
}

func (e *ExplainabilityEngine) buildStackFactor(c CandidateFeatures, p RankingParams, in ExplainInputs) (detail string, highlight string, reasons []string) {
	if in.TotalTargetCount > 0 || len(p.TargetStack) > 0 || len(p.TargetStackNames) > 0 {
		matched := in.MatchedStackCount
		total := in.TotalTargetCount
		if total <= 0 {
			total = len(p.TargetStack)
			if total <= 0 {
				total = len(p.TargetStackNames)
			}
		}

		// Format stack names if available
		var stackStr string
		if len(p.TargetStackNames) > 0 {
			stackStr = ": " + strings.Join(p.TargetStackNames, ", ")
		}

		if matched == total && total > 0 {
			detail = fmt.Sprintf("Полное совпадение по стеку (%d/%d)%s", matched, total, stackStr)
			highlight = fmt.Sprintf("полное совпадение по стеку%s", stackStr)
			reasons = append(reasons, "exact_stack_match")
		} else if matched > 0 {
			detail = fmt.Sprintf("Частичное совпадение по стеку (%d/%d)%s", matched, total, stackStr)
			if float64(matched)/float64(total) >= 0.6 {
				highlight = fmt.Sprintf("высокое совпадение по стеку (%d/%d)%s", matched, total, stackStr)
				reasons = append(reasons, "high_stack_match")
			}
		} else {
			detail = fmt.Sprintf("Совпадений по стеку не найдено (0/%d)", total)
		}
	} else {
		// No query target stack
		skillCount := len(c.CandidateStack)
		var stackStr string
		if len(c.CandidateStackNames) > 0 {
			stackStr = ": " + strings.Join(c.CandidateStackNames, ", ")
		}

		if skillCount >= 3 {
			detail = fmt.Sprintf("Широкий подтвержденный технологический стек (%d технологий)%s", skillCount, stackStr)
			highlight = fmt.Sprintf("развитый стек технологий%s", stackStr)
			reasons = append(reasons, "broad_tech_stack")
		} else {
			detail = fmt.Sprintf("Подтвержденный стек технологий%s", stackStr)
		}
	}

	return detail, highlight, reasons
}

func (e *ExplainabilityEngine) buildActivityFactor(c CandidateFeatures, in ExplainInputs) (detail string, highlight string, reasons []string) {
	var parts []string

	if c.PeriodicTasksSolved >= 3 {
		parts = append(parts, fmt.Sprintf("решено %d микро-задач платформы", c.PeriodicTasksSolved))
		reasons = append(reasons, "active_solver")
	} else if c.PeriodicTasksSolved > 0 {
		parts = append(parts, fmt.Sprintf("решено %d микро-задач", c.PeriodicTasksSolved))
	}

	if c.LastActiveAt != nil {
		diff := time.Since(*c.LastActiveAt)
		if diff <= 7*24*time.Hour {
			parts = append(parts, "активен на этой неделе")
			reasons = append(reasons, "recently_active")
		} else if diff <= 30*24*time.Hour {
			parts = append(parts, "активен в этом месяце")
		}
	}

	if len(parts) > 0 {
		detail = "Высокая активность: " + strings.Join(parts, ", ")
		if c.PeriodicTasksSolved >= 3 {
			highlight = "высокая активность на платформе"
		}
	} else {
		detail = "Стандартная активность профиля"
	}

	return detail, highlight, reasons
}
