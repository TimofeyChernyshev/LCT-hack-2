package domain

import "time"

// CategoryChangeReason - почему сменилась категория/грейд.
type CategoryChangeReason string

const (
	CategoryReasonInitialTest CategoryChangeReason = "initial_test"
	CategoryReasonRetake      CategoryChangeReason = "retake"
	CategoryReasonManual      CategoryChangeReason = "manual"
)

func (r CategoryChangeReason) Valid() bool {
	switch r {
	case CategoryReasonInitialTest, CategoryReasonRetake, CategoryReasonManual:
		return true
	}
	return false
}

// CategoryHistoryEntry - одна запись истории смены категории/грейда.
type CategoryHistoryEntry struct {
	CategoryID       string
	GradeID          string
	SpecializationID string
	Reason           string
	EffectiveFrom    time.Time
	EffectiveTo      *time.Time
}

// CategoryState - текущая категория/грейд + история изменений.
// Заполняется testing-сервисом через internal API, здесь только read-model.
type CategoryState struct {
	CategoryID       *string
	GradeID          *string
	SpecializationID *string
	History          []CategoryHistoryEntry
}

// IsAssigned - назначена ли уже категория (прошёл ли кандидат первичный тест).
func (s CategoryState) IsAssigned() bool {
	return s.CategoryID != nil && s.GradeID != nil
}
