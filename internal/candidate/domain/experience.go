package domain

import "time"

type Experience struct {
	ID          string
	UserID      string
	Company     string
	Position    string
	StartedAt   time.Time
	EndedAt     time.Time
	Description *string
	CreatedAt   time.Time
}
