package application

import (
	"context"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
)

// TransactionManager инкапсулирует работу с транзакциями.
// Позволяет application-слою описывать атомарные сценарии,
// не зная про pgx/pgxpool.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	MarkEmailVerified(ctx context.Context, userID string, at time.Time) error
	UpdatePassword(ctx context.Context, userID, hash string) error
	RecordSuccessfulLogin(ctx context.Context, userID string, at time.Time) error
	RecordFailedLogin(ctx context.Context, userID string, failedLogins int, lockedUntil *time.Time) error
	SoftDelete(ctx context.Context, userID string) error
}

type VerificationTokenRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetForUpdate(ctx context.Context, tokenHash string) (*domain.VerificationToken, error)
	MarkConsumed(ctx context.Context, id string) error
	InvalidateAll(ctx context.Context, userID string) error
}

type PasswordResetRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetForUpdate(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error)
	MarkConsumed(ctx context.Context, id string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time, ip, ua string) error
	GetForUpdate(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	MarkRevoked(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type ConsentRepository interface {
	Grant(ctx context.Context, c *domain.Consent) error
	Revoke(ctx context.Context, userID string, t domain.ConsentType) error
	ListByUser(ctx context.Context, userID string) ([]domain.Consent, error)
}

type AuditRepository interface {
	Log(ctx context.Context, userID, action, ip, ua string, payload map[string]any) error
}

type Mailer interface {
	SendEmailVerification(ctx context.Context, to, link string) error
	SendPasswordReset(ctx context.Context, to, link string) error
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash, plain string) bool
}
