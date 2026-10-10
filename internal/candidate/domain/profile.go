package domain

import "time"

type Profile struct {
	UserID           string
	FirstName        *string
	LastName         *string
	MiddleName       *string
	Headline         *string
	About            *string
	Location         *string
	YearsExperience  *float32
	CategoryID       *string
	GradeID          *string
	SpecializationID *string
	FSPMemberID      *string
	SalaryMin        *int
	SalaryMax        *int
	SalaryCurrency   *string
	SoftSkills       []string
	UpdatedAt        time.Time
}

func (p *Profile) DisplayName() string {
	if p.FirstName == nil && p.LastName == nil {
		return "Кандидат"
	}
	var s string
	if p.FirstName != nil {
		s += *p.FirstName
	}
	if p.LastName != nil {
		if s != "" {
			s += " "
		}
		s += *p.LastName
	}
	return s
}
