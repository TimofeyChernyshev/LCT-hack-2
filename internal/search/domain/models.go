package domain

import (
	"time"

	"github.com/google/uuid"

	"TimofeyChernyshev/LCT-hack-2/pkg/ranking"
)

// CandidateSearchDoc represents a candidate's indexed profile in search_db
type CandidateSearchDoc struct {
	UserID               uuid.UUID   `json:"userId"`
	DisplayName          string      `json:"displayName"`
	CategoryID           uuid.UUID   `json:"categoryId"`
	SpecializationID     uuid.UUID   `json:"specializationId"`
	SpecializationName   string      `json:"specializationName"`
	GradeID              uuid.UUID   `json:"gradeId"`
	GradeName            string      `json:"gradeName"`
	GradeRank            int         `json:"gradeRank"`
	TestScore            *float64    `json:"testScore,omitempty"`
	HasFSP               bool        `json:"hasFsp"`
	SportsRank           string      `json:"sportsRank,omitempty"`
	FSPRating            int         `json:"fspRating"`
	FSPScore             float64     `json:"fspScore"`
	FSPAchievementsCount int         `json:"fspAchievementsCount"`
	FSPBestPlace         *int        `json:"fspBestPlace,omitempty"`
	FSPWeightSum         int         `json:"fspWeightSum"`
	FSPHighlights        []string    `json:"fspHighlights,omitempty"`
	Stack                []uuid.UUID `json:"stack"`
	YearsExperience      *float64    `json:"yearsExperience,omitempty"`
	Location             *string     `json:"location,omitempty"`
	PeriodicTasksSolved  int         `json:"periodicTasksSolved"`
	ActivityScore        float64     `json:"activityScore"`
	CalculatedScore      float64     `json:"calculatedScore"`
	Explanation          string      `json:"explanation"`
	Reasons              []string    `json:"reasons"`
	LastActiveAt         time.Time   `json:"lastActiveAt"`
	UpdatedAt            time.Time   `json:"updatedAt"`
}

// SearchFilter represents query criteria from employer
type SearchFilter struct {
	CategoryID         *uuid.UUID  `json:"categoryId,omitempty"`
	SpecializationID   *uuid.UUID  `json:"specializationId,omitempty"`
	GradeID            *uuid.UUID  `json:"gradeId,omitempty"`
	Stack              []uuid.UUID `json:"stack,omitempty"`
	HasFSP             *bool       `json:"hasFsp,omitempty"`
	MinYearsExperience *float64    `json:"minYearsExperience,omitempty"`
	Location           *string     `json:"location,omitempty"`
	Sort               string      `json:"sort,omitempty"` // relevance, testScore, fspWeight
	Limit              int         `json:"limit"`
	Offset             int         `json:"offset"`
}

// SearchDocHit is a candidate result returned to the employer with scores and explanation
type SearchDocHit struct {
	UserID               uuid.UUID   `json:"userId"`
	DisplayName          *string     `json:"displayName,omitempty"`
	CategoryID           uuid.UUID   `json:"categoryId"`
	GradeID              uuid.UUID   `json:"gradeId"`
	SpecializationID     uuid.UUID   `json:"specializationId"`
	TestScore            *float32    `json:"testScore,omitempty"`
	FSPAchievementsCount *int        `json:"fspAchievementsCount,omitempty"`
	FSPBestPlace         *int        `json:"fspBestPlace,omitempty"`
	FSPWeightSum         *int        `json:"fspWeightSum,omitempty"`
	YearsExperience      *float32    `json:"yearsExperience,omitempty"`
	Score                float32     `json:"score"`
	Reasons              []string    `json:"reasons"`
	Explanation          string      `json:"explanation,omitempty"`
	Factors              []ranking.ScoreFactor `json:"factors,omitempty"`
}

// SearchPage contains paginated hits
type SearchPage struct {
	Items  []SearchDocHit `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ExplainResponse holds factor breakdown for a candidate
type ExplainResponse struct {
	UserID     uuid.UUID              `json:"userId"`
	FinalScore float32                `json:"finalScore"`
	Factors    []ranking.ScoreFactor  `json:"factors"`
	Reasons    []string               `json:"reasons"`
	Summary    string                 `json:"summary"`
}
