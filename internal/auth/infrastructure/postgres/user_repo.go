package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) *UserRepo { return &UserRepo{pool: pool} }

func (r *UserRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

const userColumns = `id, email, email_verified_at, password_hash, role, status,
                     failed_logins, locked_until, last_login_at, created_at, updated_at`

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (email, password_hash, role, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	err := r.q(ctx).QueryRow(ctx, q,
		u.Email, u.PasswordHash, string(u.Role), string(u.Status),
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE email = $1 AND status <> 'deleted'`
	return r.scanOne(ctx, q, email)
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	q := `SELECT ` + userColumns + ` FROM users WHERE id = $1 AND status <> 'deleted'`
	return r.scanOne(ctx, q, id)
}

func (r *UserRepo) scanOne(ctx context.Context, q string, args ...any) (*domain.User, error) {
	var u domain.User
	var role, status string
	err := r.q(ctx).QueryRow(ctx, q, args...).Scan(
		&u.ID, &u.Email, &u.EmailVerifiedAt, &u.PasswordHash,
		&role, &status, &u.FailedLogins, &u.LockedUntil,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.Role = domain.Role(role)
	u.Status = domain.Status(status)
	return &u, nil
}

func (r *UserRepo) MarkEmailVerified(ctx context.Context, userID string, at time.Time) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE users SET email_verified_at = $2, updated_at = now() WHERE id = $1`,
		userID, at)
	return err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, hash string) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
		userID, hash)
	return err
}

func (r *UserRepo) RecordSuccessfulLogin(ctx context.Context, userID string, at time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE users
		SET failed_logins = 0, locked_until = NULL,
		    last_login_at = $2, updated_at = now()
		WHERE id = $1`, userID, at)
	return err
}

// RecordFailedLogin — принимает уже посчитанные значения.
// Логика «когда блокировать» — на стороне application.
func (r *UserRepo) RecordFailedLogin(ctx context.Context, userID string, failedLogins int, lockedUntil *time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE users
		SET failed_logins = $2, locked_until = $3, updated_at = now()
		WHERE id = $1`, userID, failedLogins, lockedUntil)
	return err
}

func (r *UserRepo) SoftDelete(ctx context.Context, userID string) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE users
		SET status = 'deleted',
		    email = concat('deleted+', id, '@fsp.local'),
		    password_hash = '',
		    updated_at = now()
		WHERE id = $1`, userID)
	return err
}
