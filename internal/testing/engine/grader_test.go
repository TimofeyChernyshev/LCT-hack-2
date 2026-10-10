package engine

import (
	"encoding/json"
	"testing"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
)

func TestGrader_SingleChoice(t *testing.T) {
	grader := NewGrader()

	task := &domain.Task{
		Type: domain.TaskTypeSingleChoice,
	}

	item := &domain.SessionItem{
		VariantParams: json.RawMessage(`{
			"option_map": {"A": "opt_wrong", "B": "opt_correct", "C": "opt_other"},
			"correct_keys": ["B"]
		}`),
	}

	// Correct answer by letter "B"
	resB := grader.Grade(task, item, "B")
	if !resB.IsCorrect || resB.Score != 1.0 {
		t.Errorf("expected B to be correct, got isCorrect=%v, score=%v", resB.IsCorrect, resB.Score)
	}

	// Correct answer lowercase "b"
	resb := grader.Grade(task, item, "b")
	if !resb.IsCorrect || resb.Score != 1.0 {
		t.Errorf("expected lowercase b to be correct, got isCorrect=%v", resb.IsCorrect)
	}

	// Wrong answer "A"
	resA := grader.Grade(task, item, "A")
	if resA.IsCorrect || resA.Score != 0.0 {
		t.Errorf("expected A to be incorrect, got %v", resA)
	}
}

func TestGrader_MultiChoice(t *testing.T) {
	grader := NewGrader()

	task := &domain.Task{
		Type: domain.TaskTypeMultiChoice,
	}

	item := &domain.SessionItem{
		VariantParams: json.RawMessage(`{
			"option_map": {"A": "opt_1", "B": "opt_2", "C": "opt_3"},
			"correct_keys": ["A", "C"]
		}`),
	}

	// Full match
	resFull := grader.Grade(task, item, []string{"A", "C"})
	if !resFull.IsCorrect || resFull.Score != 1.0 {
		t.Errorf("expected full match to have score 1.0, got %v", resFull.Score)
	}

	// Partial match
	resPartial := grader.Grade(task, item, []string{"A"})
	if resPartial.Score != 0.5 {
		t.Errorf("expected partial match score 0.5, got %v", resPartial.Score)
	}

	// Match with penalty (A, B, C where B is wrong)
	resPenalty := grader.Grade(task, item, []string{"A", "B", "C"})
	if resPenalty.Score >= 1.0 {
		t.Errorf("expected penalty for selecting wrong option B, got %v", resPenalty.Score)
	}
}

func TestGrader_Regex(t *testing.T) {
	grader := NewGrader()

	task := &domain.Task{
		Type: domain.TaskTypeRegex,
		Solution: json.RawMessage(`{
			"matches": ["1.0.0", "2.13.4"],
			"non_matches": ["v1.0", "1.a.2"]
		}`),
	}

	item := &domain.SessionItem{}

	// Correct SemVer pattern
	resCorrect := grader.Grade(task, item, `^[0-9]+\.[0-9]+\.[0-9]+$`)
	if !resCorrect.IsCorrect || resCorrect.Score != 1.0 {
		t.Errorf("expected valid regex to pass all test cases, got score=%v", resCorrect.Score)
	}

	// Incomplete pattern (fails non_matches)
	resLoose := grader.Grade(task, item, `.*`)
	if resLoose.IsCorrect || resLoose.Score == 1.0 {
		t.Errorf("expected catch-all regex to fail, got %v", resLoose.Score)
	}

	// Invalid syntax
	resInvalid := grader.Grade(task, item, `[invalid-regex(`)
	if resInvalid.IsCorrect || resInvalid.Score != 0.0 {
		t.Errorf("expected invalid regex to return score 0, got %v", resInvalid.Score)
	}
}

func TestGrader_SQL(t *testing.T) {
	grader := NewGrader()

	task := &domain.Task{
		Type: domain.TaskTypeSQL,
		Solution: json.RawMessage(`{
			"required_keywords": ["SELECT", "FROM", "WHERE", "ORDER BY"],
			"target_tables": ["users"]
		}`),
	}

	item := &domain.SessionItem{}

	resValid := grader.Grade(task, item, "SELECT id, email FROM users WHERE active = true ORDER BY created_at DESC;")
	if !resValid.IsCorrect || resValid.Score < 0.85 {
		t.Errorf("expected valid SQL to pass, got score=%v", resValid.Score)
	}

	resDrop := grader.Grade(task, item, "DROP TABLE users; SELECT 1;")
	if resDrop.IsCorrect || resDrop.Score != 0.0 {
		t.Errorf("expected destructive SQL with DROP to be rejected, got %v", resDrop)
	}
}

func TestGrader_Code(t *testing.T) {
	grader := NewGrader()

	task := &domain.Task{
		Type: domain.TaskTypeCode,
		Solution: json.RawMessage(`{
			"required_patterns": ["sync.Mutex", "Lock()", "Unlock()"],
			"forbidden_tokens": ["panic"]
		}`),
	}

	item := &domain.SessionItem{}

	codeValid := `
		type SafeCounter struct {
			mu sync.Mutex
			v int
		}
		func (c *SafeCounter) Inc() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.v++
		}
	`
	res := grader.Grade(task, item, codeValid)
	if !res.IsCorrect || res.Score < 0.8 {
		t.Errorf("expected valid code to pass, got isCorrect=%v, score=%v", res.IsCorrect, res.Score)
	}

	codePanic := `func (c *SafeCounter) Inc() { panic("fail") }`
	resPanic := grader.Grade(task, item, codePanic)
	if resPanic.IsCorrect || resPanic.Score != 0.0 {
		t.Errorf("expected forbidden token panic to be rejected")
	}
}
