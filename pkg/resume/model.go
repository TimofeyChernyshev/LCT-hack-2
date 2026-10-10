package resume

import (
	"time"

	"github.com/google/uuid"
)

// ParsedResume represents structured data extracted from candidate resume (PDF or text)
type ParsedResume struct {
	FullName             string            `json:"fullName"`
	FirstName            string            `json:"firstName,omitempty"`
	LastName             string            `json:"lastName,omitempty"`
	MiddleName           string            `json:"middleName,omitempty"`
	Headline             string            `json:"headline"` // e.g. "Senior Backend Developer (Go)"
	Specialization       string            `json:"specialization,omitempty"`
	SuggestedGrade       string            `json:"suggestedGrade,omitempty"` // Junior, Middle, Senior, Lead
	Location             string            `json:"location,omitempty"`
	Contacts             ExtractedContacts `json:"contacts"`
	TotalYearsExperience float64           `json:"totalYearsExperience"`
	Skills               []string          `json:"skills"` // e.g. ["Go", "PostgreSQL", "Docker", "Redis"]
	Experiences          []ExperienceItem  `json:"experiences"`
	Education            []EducationItem   `json:"education"`
	FSPMention           *FSPMention       `json:"fspMention,omitempty"`
	RawTextPreview       string            `json:"rawTextPreview,omitempty"`
}

// ExtractedContacts contains all communication channels found in resume
type ExtractedContacts struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Telegram string `json:"telegram,omitempty"`
	GitHub   string `json:"github,omitempty"`
	LinkedIn string `json:"linkedin,omitempty"`
	Website  string `json:"website,omitempty"`
}

// ExperienceItem represents a single work period
type ExperienceItem struct {
	Company          string   `json:"company"`
	Position         string   `json:"position"`
	Period           string   `json:"period,omitempty"`
	StartedAt        string   `json:"startedAt,omitempty"` // YYYY-MM
	EndedAt          string   `json:"endedAt,omitempty"`   // YYYY-MM or "настоящее время"
	DurationYears    float64  `json:"durationYears,omitempty"`
	Description      string   `json:"description,omitempty"`
	TechnologiesUsed []string `json:"technologiesUsed,omitempty"`
}

// EducationItem represents university or training degree
type EducationItem struct {
	Institution    string `json:"institution"`
	Degree         string `json:"degree,omitempty"`
	FieldOfStudy   string `json:"fieldOfStudy,omitempty"`
	GraduationYear int    `json:"graduationYear,omitempty"`
}

// FSPMention holds any mentions of sports programming achievements found in resume
type FSPMention struct {
	HasFSP       bool     `json:"hasFsp"`
	FSPID        string   `json:"fspId,omitempty"`
	SportsRank   string   `json:"sportsRank,omitempty"` // ЗМС, МС, КМС, разряды
	Rating       int      `json:"rating,omitempty"`
	Achievements []string `json:"achievements,omitempty"`
}

// CandidateExportProfile represents inputs for generating a standardized branded PDF profile
type CandidateExportProfile struct {
	UserID             uuid.UUID `json:"userId"`
	FullName           string    `json:"fullName"`
	Headline           string    `json:"headline,omitempty"`
	SpecializationName string    `json:"specializationName,omitempty"`
	GradeName          string    `json:"gradeName,omitempty"`
	GradeRank          int       `json:"gradeRank,omitempty"`
	Location           string    `json:"location,omitempty"`
	TestScore          float64   `json:"testScore,omitempty"`
	TestPercentile     float64   `json:"testPercentile,omitempty"`
	HasFSP             bool      `json:"hasFsp,omitempty"`
	SportsRank         string    `json:"sportsRank,omitempty"`
	FSPRating          int       `json:"fspRating,omitempty"`
	FSPAchievements    []string  `json:"fspAchievements,omitempty"`
	Skills             []string  `json:"skills,omitempty"`
	YearsExperience    float64   `json:"yearsExperience,omitempty"`
	MaskContacts       bool      `json:"maskContacts"` // true if employer has not accepted invitation yet (152-FZ privacy)
	Email              string    `json:"email,omitempty"`
	Phone              string    `json:"phone,omitempty"`
	Telegram           string    `json:"telegram,omitempty"`
	GitHub             string    `json:"github,omitempty"`
	GeneratedAt        time.Time `json:"generatedAt,omitempty"`
}
