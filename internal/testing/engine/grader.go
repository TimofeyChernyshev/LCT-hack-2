package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
)

type Grader struct{}

func NewGrader() *Grader {
	return &Grader{}
}

type GradeResult struct {
	IsCorrect bool
	Score     float64
	GradedBy  string
	Feedback  string
}

// Grade evaluates the submitted answer for a session item
func (g *Grader) Grade(task *domain.Task, item *domain.SessionItem, rawAnswer interface{}) GradeResult {
	if rawAnswer == nil {
		return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto", Feedback: "empty answer"}
	}

	var variant VariantParams
	_ = json.Unmarshal(item.VariantParams, &variant)

	switch task.Type {
	case domain.TaskTypeSingleChoice:
		return g.gradeSingleChoice(variant, rawAnswer)
	case domain.TaskTypeMultiChoice:
		return g.gradeMultiChoice(variant, rawAnswer)
	case domain.TaskTypeText:
		return g.gradeText(task, variant, rawAnswer)
	case domain.TaskTypeRegex:
		return g.gradeRegex(task, rawAnswer)
	case domain.TaskTypeSQL:
		return g.gradeSQL(task, rawAnswer)
	case domain.TaskTypeCode:
		return g.gradeCode(task, variant, rawAnswer)
	default:
		return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto"}
	}
}

func (g *Grader) gradeSingleChoice(variant VariantParams, rawAnswer interface{}) GradeResult {
	userAnswer := strings.TrimSpace(fmt.Sprint(rawAnswer))
	userAnswer = strings.ToUpper(userAnswer)

	// Check if matches correct key directly (e.g. "A" or "B")
	for _, k := range variant.CorrectKeys {
		if strings.EqualFold(userAnswer, k) {
			return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
		}
	}

	// Check if user passed the original option ID
	if optID, exists := variant.OptionMap[userAnswer]; exists {
		for _, k := range variant.CorrectKeys {
			if variant.OptionMap[k] == optID {
				return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
			}
		}
	}

	return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto"}
}

func (g *Grader) gradeMultiChoice(variant VariantParams, rawAnswer interface{}) GradeResult {
	var userKeys []string
	switch v := rawAnswer.(type) {
	case []interface{}:
		for _, item := range v {
			userKeys = append(userKeys, strings.ToUpper(strings.TrimSpace(fmt.Sprint(item))))
		}
	case []string:
		for _, item := range v {
			userKeys = append(userKeys, strings.ToUpper(strings.TrimSpace(item)))
		}
	case string:
		parts := strings.Split(v, ",")
		for _, p := range parts {
			if s := strings.ToUpper(strings.TrimSpace(p)); s != "" {
				userKeys = append(userKeys, s)
			}
		}
	default:
		userKeys = []string{strings.ToUpper(strings.TrimSpace(fmt.Sprint(rawAnswer)))}
	}

	correctSet := make(map[string]bool)
	for _, k := range variant.CorrectKeys {
		correctSet[strings.ToUpper(k)] = true
	}

	userSet := make(map[string]bool)
	for _, k := range userKeys {
		userSet[k] = true
	}

	if len(correctSet) == 0 {
		return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto"}
	}

	correctCount := 0
	for k := range userSet {
		if correctSet[k] {
			correctCount++
		} else {
			// Penalty for wrong option
			correctCount--
		}
	}

	if correctCount < 0 {
		correctCount = 0
	}

	score := float64(correctCount) / float64(len(correctSet))
	if score > 1.0 {
		score = 1.0
	}

	isCorrect := len(userSet) == len(correctSet) && correctCount == len(correctSet)
	return GradeResult{
		IsCorrect: isCorrect,
		Score:     score,
		GradedBy:  "auto",
	}
}

func (g *Grader) gradeText(task *domain.Task, variant VariantParams, rawAnswer interface{}) GradeResult {
	ans := strings.TrimSpace(strings.ToLower(fmt.Sprint(rawAnswer)))

	// Check dynamic expected value first
	if variant.ExpectedValue != "" {
		expected := strings.TrimSpace(strings.ToLower(variant.ExpectedValue))
		if ans == expected {
			return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
		}
	}

	// Check task solution keywords
	if len(task.Solution) > 0 {
		var sol struct {
			Expected string   `json:"expected"`
			Keywords []string `json:"keywords"`
		}
		if err := json.Unmarshal(task.Solution, &sol); err == nil {
			if sol.Expected != "" && ans == strings.TrimSpace(strings.ToLower(sol.Expected)) {
				return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
			}
			if len(sol.Keywords) > 0 {
				matched := 0
				for _, kw := range sol.Keywords {
					if strings.Contains(ans, strings.ToLower(kw)) {
						matched++
					}
				}
				score := float64(matched) / float64(len(sol.Keywords))
				return GradeResult{
					IsCorrect: score >= 0.8,
					Score:     score,
					GradedBy:  "auto",
				}
			}
		}
	}

	return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto"}
}

func (g *Grader) gradeRegex(task *domain.Task, rawAnswer interface{}) GradeResult {
	pattern := strings.TrimSpace(fmt.Sprint(rawAnswer))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto", Feedback: "invalid regex syntax"}
	}

	var sol struct {
		Matches    []string `json:"matches"`
		NonMatches []string `json:"non_matches"`
	}
	if err := json.Unmarshal(task.Solution, &sol); err != nil {
		return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
	}

	total := len(sol.Matches) + len(sol.NonMatches)
	if total == 0 {
		return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
	}

	passed := 0
	for _, m := range sol.Matches {
		if re.MatchString(m) {
			passed++
		}
	}
	for _, nm := range sol.NonMatches {
		if !re.MatchString(nm) {
			passed++
		}
	}

	score := float64(passed) / float64(total)
	return GradeResult{
		IsCorrect: score >= 0.99,
		Score:     score,
		GradedBy:  "auto",
	}
}

func (g *Grader) gradeSQL(task *domain.Task, rawAnswer interface{}) GradeResult {
	query := strings.ToUpper(strings.TrimSpace(fmt.Sprint(rawAnswer)))

	// Check forbidden destructive statements
	for _, forbidden := range []string{"DROP ", "TRUNCATE ", "DELETE "} {
		if strings.Contains(query, forbidden) {
			return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto", Feedback: "forbidden SQL keyword"}
		}
	}

	var sol struct {
		RequiredKeywords []string `json:"required_keywords"`
		TargetTables     []string `json:"target_tables"`
	}
	if err := json.Unmarshal(task.Solution, &sol); err != nil {
		return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
	}

	totalChecks := len(sol.RequiredKeywords) + len(sol.TargetTables)
	if totalChecks == 0 {
		return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
	}

	passed := 0
	for _, kw := range sol.RequiredKeywords {
		if strings.Contains(query, strings.ToUpper(kw)) {
			passed++
		}
	}
	for _, tbl := range sol.TargetTables {
		if strings.Contains(query, strings.ToUpper(tbl)) {
			passed++
		}
	}

	score := float64(passed) / float64(totalChecks)
	return GradeResult{
		IsCorrect: score >= 0.85,
		Score:     score,
		GradedBy:  "auto",
	}
}

func (g *Grader) gradeCode(task *domain.Task, variant VariantParams, rawAnswer interface{}) GradeResult {
	code := strings.TrimSpace(fmt.Sprint(rawAnswer))
	if len(code) == 0 {
		return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto", Feedback: "empty code"}
	}

	var sol struct {
		RequiredPatterns []string          `json:"required_patterns"`
		ForbiddenTokens  []string          `json:"forbidden_tokens"`
		TestCases        []TestCasePattern `json:"test_cases"`
	}
	_ = json.Unmarshal(task.Solution, &sol)

	for _, token := range sol.ForbiddenTokens {
		if strings.Contains(code, token) {
			return GradeResult{IsCorrect: false, Score: 0.0, GradedBy: "auto", Feedback: "forbidden token in code"}
		}
	}

	totalChecks := len(sol.RequiredPatterns) + len(sol.TestCases)
	if totalChecks == 0 {
		// Non-empty code passes baseline rubric
		return GradeResult{IsCorrect: true, Score: 1.0, GradedBy: "auto"}
	}

	passed := 0
	for _, p := range sol.RequiredPatterns {
		if re, err := regexp.Compile(p); err == nil && re.MatchString(code) {
			passed++
		} else if strings.Contains(code, p) {
			passed++
		}
	}

	score := float64(passed) / float64(totalChecks)
	if len(sol.RequiredPatterns) > 0 && passed == len(sol.RequiredPatterns) {
		score = 1.0
	}

	return GradeResult{
		IsCorrect: score >= 0.8,
		Score:     score,
		GradedBy:  "auto",
	}
}

type TestCasePattern struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}
