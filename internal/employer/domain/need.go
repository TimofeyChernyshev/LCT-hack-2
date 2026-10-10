package domain

import "time"

type NeedStatus string

const (
	NeedStatusDraft  NeedStatus = "draft"
	NeedStatusActive NeedStatus = "active"
	NeedStatusClosed NeedStatus = "closed"
)

type WorkFormat string

const (
	WorkFormatRemote WorkFormat = "remote"
	WorkFormatHybrid WorkFormat = "hybrid"
	WorkFormatOffice WorkFormat = "office"
)

type Need struct {
	ID                 string
	CompanyID          string
	Title              string
	Description        string
	CategoryID         *string
	SpecializationID   *string
	GradeID            *string
	Stack              []string
	SalaryMin          int
	SalaryMax          int
	Currency           string
	WorkFormat         *WorkFormat
	Location           *string
	MinExperienceYears *float32
	Status             NeedStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
