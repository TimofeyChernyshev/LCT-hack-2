package application

import (
	"context"
	"io"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	Create(ctx context.Context, p *domain.Profile) error
	Update(ctx context.Context, p *domain.Profile) error
}

type ContactsRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Contacts, error)
	Upsert(ctx context.Context, userID string, c *domain.Contacts) error
}

type VisibilityRepository interface {
	Get(ctx context.Context, userID string) (*domain.Visibility, error)
	Upsert(ctx context.Context, userID string, v *domain.Visibility) error
}

type RevealRepository interface {
	// IsRevealed — есть ли accepted-приглашение или отклик между парой.
	IsRevealed(ctx context.Context, candidateUserID, employerUserID string) (bool, error)
	Create(ctx context.Context, r *domain.ContactReveal) error
}

type ResumeRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Resume, error)
	Create(ctx context.Context, r *domain.Resume) error
	GetByID(ctx context.Context, id string) (*domain.Resume, error)
	UnsetPrimary(ctx context.Context, userID string) error
}

type ExperienceRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.Experience, error)
	Create(ctx context.Context, e *domain.Experience) error
}

type TechnologyRepository interface {
	ReplaceForUser(ctx context.Context, userID string, techs []domain.CandidateTechnology) error
	ListByUser(ctx context.Context, userID string) ([]domain.CandidateTechnology, error)
}

type FSPRepository interface {
	GetMemberID(ctx context.Context, userID string) (*string, *time.Time, error)
	LinkMemberID(ctx context.Context, userID, memberID string) error
	ListAchievements(ctx context.Context, userID string) ([]domain.FSPAchievement, error)
	CountAchievements(ctx context.Context, userID string) (int, *int, error)
	CreateAchievement(ctx context.Context, a *domain.FSPAchievement) error
}

type CategoryRepository interface {
	GetState(ctx context.Context, userID string) (*domain.CategoryState, error)
}

type FileStorage interface {
	Save(ctx context.Context, userID, filename string, r io.Reader, maxBytes int64) (path string, err error)
}
