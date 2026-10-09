package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusSubmitted  SessionStatus = "submitted"
	SessionStatusEvaluated  SessionStatus = "evaluated"
	SessionStatusExpired    SessionStatus = "expired"
	SessionStatusCancelled  SessionStatus = "cancelled"
)

type SessionItemStatus string

const (
	SessionItemPending  SessionItemStatus = "pending"
	SessionItemAnswered SessionItemStatus = "answered"
	SessionItemSkipped  SessionItemStatus = "skipped"
)

type TaskType string

const (
	TaskTypeSingleChoice TaskType = "single_choice"
	TaskTypeMultiChoice  TaskType = "multi_choice"
	TaskTypeCode         TaskType = "code"
	TaskTypeText         TaskType = "text"
	TaskTypeSQL          TaskType = "sql"
	TaskTypeRegex        TaskType = "regex"
)

type TemplateConfig struct {
	ItemsPerSession   int     `json:"items_per_session,omitempty"`
	DurationMinutes   int     `json:"duration_minutes,omitempty"`
	MinScoreToPass    float64 `json:"min_score_to_pass,omitempty"`
	MinScoreToUpgrade float64 `json:"min_score_to_upgrade,omitempty"`
}

type Template struct {
	ID         uuid.UUID       `json:"id"`
	CategoryID uuid.UUID       `json:"category_id"`
	Version    int             `json:"version"`
	Config     json.RawMessage `json:"config"`
	IsActive   bool            `json:"is_active"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Task struct {
	ID             uuid.UUID       `json:"id"`
	TemplateID     uuid.UUID       `json:"template_id"`
	Type           TaskType        `json:"type"`
	Topic          string          `json:"topic"`
	Title          string          `json:"title"`
	Body           string          `json:"body"`
	Generator      json.RawMessage `json:"generator,omitempty"`
	Solution       json.RawMessage `json:"solution,omitempty"`
	Rubric         json.RawMessage `json:"rubric,omitempty"`
	Difficulty     int             `json:"difficulty"`      // 1..5
	Discrimination float64         `json:"discrimination"`  // IRT a
	DifficultyIRT  float64         `json:"difficulty_irt"`  // IRT b
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Session struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	TargetCategoryID    uuid.UUID       `json:"target_category_id"`
	TemplateID          uuid.UUID       `json:"template_id"`
	Status              SessionStatus   `json:"status"`
	AbilityEstimate     *float64        `json:"ability_estimate,omitempty"`
	Score               *float64        `json:"score,omitempty"`
	ResultingGradeID    *uuid.UUID      `json:"resulting_grade_id,omitempty"`
	ResultingCategoryID *uuid.UUID      `json:"resulting_category_id,omitempty"`
	StartedAt           time.Time       `json:"started_at"`
	FinishedAt          *time.Time      `json:"finished_at,omitempty"`
	Items               []SessionItem   `json:"items,omitempty"`
}

type SessionItem struct {
	ID            uuid.UUID         `json:"id"`
	SessionID     uuid.UUID         `json:"session_id"`
	TaskID        uuid.UUID         `json:"task_id"`
	Position      int               `json:"position"`
	VariantParams json.RawMessage   `json:"variant_params"`
	RenderedBody  string            `json:"rendered_body"`
	Status        SessionItemStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`

	// Attached task details (not exposed directly to user via rendered view)
	Task *Task `json:"task,omitempty"`
}

type TestAnswer struct {
	ID         uuid.UUID       `json:"id"`
	ItemID     uuid.UUID       `json:"item_id"`
	Answer     json.RawMessage `json:"answer"`
	IsCorrect  *bool           `json:"is_correct,omitempty"`
	Score      *float64        `json:"score,omitempty"`
	GradedBy   string          `json:"graded_by"`
	GradedAt   *time.Time      `json:"graded_at,omitempty"`
	AnsweredAt time.Time       `json:"answered_at"`
}

type GradeChangeEvent struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	FromGradeID *uuid.UUID `json:"from_grade_id,omitempty"`
	ToGradeID   uuid.UUID  `json:"to_grade_id"`
	Reason      string     `json:"reason"`
	ChangedAt   time.Time  `json:"changed_at"`
}

type PeriodicTask struct {
	ID             uuid.UUID `json:"id"`
	EmployerUserID uuid.UUID `json:"employer_user_id"`
	CategoryID     uuid.UUID `json:"category_id"`
	Title          string    `json:"title"`
	Body           string    `json:"body"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type PeriodicSubmission struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Answer    string    `json:"answer"`
	Score     *float64  `json:"score,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
