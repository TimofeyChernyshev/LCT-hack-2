package engine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
)

type GeneratorRule struct {
	Type     string        `json:"type"` // "int_range", "choice", "array_shuffle", "math"
	Min      int           `json:"min,omitempty"`
	Max      int           `json:"max,omitempty"`
	Choices  []string      `json:"choices,omitempty"`
	Elements []interface{} `json:"elements,omitempty"`
	Formula  string        `json:"formula,omitempty"` // e.g. "a % b", "a * b"
}

type GeneratorConfig struct {
	Variables map[string]GeneratorRule `json:"variables,omitempty"`
	Options   []OptionDef              `json:"options,omitempty"`
	Formula   string                   `json:"formula,omitempty"`
}

type OptionDef struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct,omitempty"`
}

type VariantParams struct {
	GeneratedVars map[string]interface{} `json:"vars,omitempty"`
	OptionMap     map[string]string      `json:"option_map,omitempty"` // Presentation key -> Option ID
	CorrectKeys   []string               `json:"correct_keys,omitempty"`
	ExpectedValue string                 `json:"expected,omitempty"`
}

type TaskGenerator struct {
	rng *rand.Rand
}

func NewTaskGenerator() *TaskGenerator {
	return &TaskGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateItem compiles a template task into a unique SessionItem with randomized variables and shuffled options
func (g *TaskGenerator) GenerateItem(task *domain.Task, sessionID string, position int) (*domain.SessionItem, error) {
	variant := VariantParams{
		GeneratedVars: make(map[string]interface{}),
		OptionMap:     make(map[string]string),
	}

	body := task.Body

	// 1. Parse Generator Config if present
	if len(task.Generator) > 0 && string(task.Generator) != "{}" && string(task.Generator) != "null" {
		var genConfig GeneratorConfig
		if err := json.Unmarshal(task.Generator, &genConfig); err == nil {
			// Generate variables
			for varName, rule := range genConfig.Variables {
				val := g.generateVariable(rule, variant.GeneratedVars)
				variant.GeneratedVars[varName] = val
				body = strings.ReplaceAll(body, fmt.Sprintf("{{%s}}", varName), fmt.Sprint(val))
			}

			// If formula is defined for expected answer
			if genConfig.Formula != "" {
				computed := g.evaluateFormula(genConfig.Formula, variant.GeneratedVars)
				variant.ExpectedValue = computed
			}

			// If options are defined in generator
			if len(genConfig.Options) > 0 {
				body, variant = g.shuffleOptions(body, genConfig.Options, variant)
			}
		}
	}

	// 2. If options are in task.Solution or task.Body as JSON
	if len(variant.OptionMap) == 0 && (task.Type == domain.TaskTypeSingleChoice || task.Type == domain.TaskTypeMultiChoice) {
		var options []OptionDef
		if len(task.Solution) > 0 {
			var sol struct {
				Options []OptionDef `json:"options"`
				Correct []string    `json:"correct"`
			}
			if err := json.Unmarshal(task.Solution, &sol); err == nil && len(sol.Options) > 0 {
				for i := range sol.Options {
					for _, c := range sol.Correct {
						if sol.Options[i].ID == c {
							sol.Options[i].IsCorrect = true
						}
					}
				}
				options = sol.Options
			}
		}

		if len(options) > 0 {
			body, variant = g.shuffleOptions(body, options, variant)
		}
	}

	// 3. Fallback expected value from task.Solution if not dynamically set
	if variant.ExpectedValue == "" && len(task.Solution) > 0 {
		var sol struct {
			Expected string   `json:"expected"`
			Correct  []string `json:"correct"`
		}
		if err := json.Unmarshal(task.Solution, &sol); err == nil {
			if sol.Expected != "" {
				variant.ExpectedValue = sol.Expected
			}
		}
	}

	variantJSON, err := json.Marshal(variant)
	if err != nil {
		variantJSON = []byte("{}")
	}

	return &domain.SessionItem{
		Position:      position,
		VariantParams: variantJSON,
		RenderedBody:  body,
		Status:        domain.SessionItemPending,
		CreatedAt:     time.Now(),
		Task:          task,
	}, nil
}

func (g *TaskGenerator) generateVariable(rule GeneratorRule, currentVars map[string]interface{}) interface{} {
	switch rule.Type {
	case "int_range":
		if rule.Max > rule.Min {
			return rule.Min + g.rng.Intn(rule.Max-rule.Min+1)
		}
		return rule.Min
	case "choice":
		if len(rule.Choices) > 0 {
			return rule.Choices[g.rng.Intn(len(rule.Choices))]
		}
		return ""
	case "array_shuffle":
		if len(rule.Elements) > 0 {
			arr := make([]interface{}, len(rule.Elements))
			copy(arr, rule.Elements)
			g.rng.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
			return arr
		}
		return []interface{}{}
	case "math":
		return g.evaluateFormula(rule.Formula, currentVars)
	default:
		return ""
	}
}

func (g *TaskGenerator) evaluateFormula(formula string, vars map[string]interface{}) string {
	formula = strings.TrimSpace(formula)
	parts := strings.Split(formula, " ")
	if len(parts) == 3 {
		leftStr := fmt.Sprint(vars[parts[0]])
		op := parts[1]
		rightStr := fmt.Sprint(vars[parts[2]])

		left, err1 := strconv.Atoi(leftStr)
		right, err2 := strconv.Atoi(rightStr)
		if err1 == nil && err2 == nil {
			switch op {
			case "+":
				return strconv.Itoa(left + right)
			case "-":
				return strconv.Itoa(left - right)
			case "*":
				return strconv.Itoa(left * right)
			case "/":
				if right != 0 {
					return strconv.Itoa(left / right)
				}
			case "%":
				if right != 0 {
					return strconv.Itoa(left % right)
				}
			}
		}
	}
	return formula
}

func (g *TaskGenerator) shuffleOptions(currentBody string, options []OptionDef, variant VariantParams) (string, VariantParams) {
	shuffled := make([]OptionDef, len(options))
	copy(shuffled, options)
	g.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	labels := []string{"A", "B", "C", "D", "E", "F", "G"}
	var optionsText strings.Builder
	optionsText.WriteString("\n\n**Варианты ответа:**\n")

	variant.CorrectKeys = []string{}
	for i, opt := range shuffled {
		label := fmt.Sprintf("%d", i+1)
		if i < len(labels) {
			label = labels[i]
		}
		variant.OptionMap[label] = opt.ID
		if opt.IsCorrect {
			variant.CorrectKeys = append(variant.CorrectKeys, label)
		}
		optionsText.WriteString(fmt.Sprintf("- **%s**: %s\n", label, opt.Text))
	}

	rendered := currentBody + optionsText.String()
	return rendered, variant
}
