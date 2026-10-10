package domain

import "time"

type VacancyStatus string

const (
	VacancyStatusDraft     VacancyStatus = "draft"
	VacancyStatusPublished VacancyStatus = "published"
	VacancyStatusArchived  VacancyStatus = "archived"
)

type Vacancy struct {
	ID               string
	CompanyID        string
	NeedID           *string
	Title            string
	Description      string
	CategoryID       *string
	SpecializationID *string
	GradeID          *string
	Stack            []string
	SalaryMin        int
	SalaryMax        int
	Currency         string
	WorkFormat       *WorkFormat
	Location         *string
	Status           VacancyStatus
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
