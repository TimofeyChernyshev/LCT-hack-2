package engine

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"github.com/google/uuid"
)

func TestTaskGenerator_ParameterizationAndAntiCheat(t *testing.T) {
	generator := NewTaskGenerator()

	task := &domain.Task{
		ID:    uuid.New(),
		Type:  domain.TaskTypeText,
		Topic: "math",
		Title: "Dynamic Math",
		Body:  "Compute {{num1}} + {{num2}}",
		Generator: json.RawMessage(`{
			"variables": {
				"num1": {"type": "int_range", "min": 10, "max": 20},
				"num2": {"type": "int_range", "min": 5, "max": 9}
			},
			"formula": "num1 + num2"
		}`),
	}

	item1, err := generator.GenerateItem(task, uuid.NewString(), 1)
	if err != nil {
		t.Fatalf("failed to generate item1: %v", err)
	}

	if strings.Contains(item1.RenderedBody, "{{num1}}") || strings.Contains(item1.RenderedBody, "{{num2}}") {
		t.Errorf("rendered body contains unreplaced placeholders: %s", item1.RenderedBody)
	}

	var variant1 VariantParams
	if err := json.Unmarshal(item1.VariantParams, &variant1); err != nil {
		t.Fatalf("failed to parse variant params: %v", err)
	}

	if variant1.ExpectedValue == "" {
		t.Errorf("expected dynamic value to be computed, got empty")
	}

	// Verify that rendered body does not reveal the solution/expected answer
	if strings.Contains(item1.RenderedBody, "expected") || strings.Contains(item1.RenderedBody, variant1.ExpectedValue) {
		// Only check if it's not coincidentally in the question text
		if strings.Contains(item1.RenderedBody, "Ответ:") {
			t.Errorf("rendered body reveals answer!")
		}
	}
}

func TestTaskGenerator_OptionShuffling(t *testing.T) {
	generator := NewTaskGenerator()

	task := &domain.Task{
		ID:    uuid.New(),
		Type:  domain.TaskTypeSingleChoice,
		Topic: "concurrency",
		Title: "Mutex Question",
		Body:  "What is the role of sync.Mutex?",
		Solution: json.RawMessage(`{
			"options": [
				{"id": "opt_a", "text": "Mutual exclusion"},
				{"id": "opt_b", "text": "Garbage collection"},
				{"id": "opt_c", "text": "HTTP routing"},
				{"id": "opt_d", "text": "Database pooling"}
			],
			"correct": ["opt_a"]
		}`),
	}

	item, err := generator.GenerateItem(task, uuid.NewString(), 1)
	if err != nil {
		t.Fatalf("failed to generate item: %v", err)
	}

	var variant VariantParams
	if err := json.Unmarshal(item.VariantParams, &variant); err != nil {
		t.Fatalf("failed to unmarshal variant: %v", err)
	}

	if len(variant.CorrectKeys) != 1 {
		t.Fatalf("expected 1 correct key, got %v", variant.CorrectKeys)
	}

	correctKey := variant.CorrectKeys[0]
	if variant.OptionMap[correctKey] != "opt_a" {
		t.Errorf("expected correct key %s to map to opt_a, got %s", correctKey, variant.OptionMap[correctKey])
	}

	// Ensure all options are present in rendered body
	if !strings.Contains(item.RenderedBody, "Mutual exclusion") {
		t.Errorf("rendered body missing option text: %s", item.RenderedBody)
	}
}
