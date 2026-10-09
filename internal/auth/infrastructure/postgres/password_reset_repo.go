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

type PasswordResetRepo struct{ pool *pgxpool.Pool }

func NewPasswordResetRepo(pool *pgxpool.Pool) *PasswordResetRepo {
	return &PasswordResetRepo{pool: pool}
}

func (r *PasswordResetRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *PasswordResetRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

func (r *PasswordResetRepo) GetForUpdate(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var t domain.PasswordResetToken
	var consumedAt *time.Time
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, user_id, consumed_at, expires_at
		FROM password_resets
		WHERE token_hash = $1
		FOR UPDATE`, tokenHash).
		Scan(&t.ID, &t.UserID, &consumedAt, &t.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	t.Consumed = consumedAt != nil
	return &t, nil
}

func (r *PasswordResetRepo) MarkConsumed(ctx context.Context, id string) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE password_resets SET consumed_at = now() WHERE id = $1`, id)
	return err
}
