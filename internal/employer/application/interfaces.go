package application

import (
	"context"
	"errors"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
)

var ErrNotFound = errors.New("not found")

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type CompanyRepository interface {
	GetByOwner(ctx context.Context, ownerID string) (*domain.Company, error)
	Create(ctx context.Context, c *domain.Company) error
	Update(ctx context.Context, c *domain.Company) error
}

type NeedRepository interface {
	ListByCompany(ctx context.Context, companyID string) ([]domain.Need, error)
	GetByID(ctx context.Context, id string) (*domain.Need, error)
	Create(ctx context.Context, n *domain.Need) error
	Update(ctx context.Context, n *domain.Need) error
	Delete(ctx context.Context, id, companyID string) error
}

type VacancyRepository interface {
	ListByCompany(ctx context.Context, companyID string) ([]domain.Vacancy, error)
	ListPublic(ctx context.Context, filter PublicVacancyFilter) ([]domain.Vacancy, int, error)
	GetByID(ctx context.Context, id string) (*domain.Vacancy, error)
	Create(ctx context.Context, v *domain.Vacancy) error
	Update(ctx context.Context, v *domain.Vacancy) error
	Delete(ctx context.Context, id, companyID string) error
	Publish(ctx context.Context, id, companyID string) error
}

type PublicVacancyFilter struct {
	CategoryID *string
	GradeID    *string
	Limit      int
	Offset     int
}

// SearchClient — прокси в search-сервис.
type SearchClient interface {
	SearchCandidates(ctx context.Context, authHeader string, q SearchQuery) (*SearchPage, error)
	MatchesForNeed(ctx context.Context, authHeader, needID string, q MatchesQuery) (*SearchPage, error)
}

// CandidateClient — прокси в candidate-сервис.
type CandidateClient interface {
	GetCard(ctx context.Context, authHeader, candidateUserID string) (*CandidateCard, error)
}

type SearchQuery struct {
	CategoryID         *string
	SpecializationID   *string
	GradeID            *string
	Stack              []string
	HasFSP             *bool
	MinYearsExperience *float32
	Location           *string
	Sort               string
	Limit              int
	Offset             int
}

type MatchesQuery struct {
	Limit  int
	Offset int
}

type SearchPage struct {
	Items  []SearchHit
	Total  int
	Limit  int
	Offset int
}

type SearchHit struct {
	UserID               string
	DisplayName          string
	CategoryID           string
	GradeID              string
	SpecializationID     string
	TestScore            *float32
	FSPAchievementsCount int
	FSPBestPlace         *int
	FSPWeightSum         int
	YearsExperience      *float32
	Score                float32
	Reasons              []string
}

type CandidateCard struct {
	UserID           string
	DisplayName      string
	Headline         *string
	CategoryID       *string
	GradeID          *string
	SpecializationID *string
	YearsExperience  *float32
	Location         *string
	SoftSkills       []string
	Salary           *CardSalary
	FSP              *CardFSP
	Contacts         *CardContacts
}

type CardSalary struct {
	Min      *int
	Max      *int
	Currency *string
	Masked   bool
}

type CardFSP struct {
	HasFSP            bool
	AchievementsCount int
	BestPlace         *int
}

type CardContacts struct {
	Email        *string
	Phone        *string
	Telegram     *string
	GitHub       *string
	LinkedIn     *string
	Website      *string
	Masked       bool
	MaskedFields []string
}
