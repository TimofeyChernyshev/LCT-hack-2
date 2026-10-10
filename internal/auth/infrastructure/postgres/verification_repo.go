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

type VerificationTokenRepo struct{ pool *pgxpool.Pool }

func NewVerificationTokenRepo(pool *pgxpool.Pool) *VerificationTokenRepo {
	return &VerificationTokenRepo{pool: pool}
}

func (r *VerificationTokenRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *VerificationTokenRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO email_verifications (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, userID, tokenHash, expiresAt)
	return err
}

func (r *VerificationTokenRepo) GetForUpdate(ctx context.Context, tokenHash string) (*domain.VerificationToken, error) {
	var t domain.VerificationToken
	var consumedAt *time.Time
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, user_id, consumed_at, expires_at
		FROM email_verifications
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

func (r *VerificationTokenRepo) MarkConsumed(ctx context.Context, id string) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE email_verifications SET consumed_at = now() WHERE id = $1`, id)
	return err
}

func (r *VerificationTokenRepo) InvalidateAll(ctx context.Context, userID string) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE email_verifications
		SET consumed_at = now()
		WHERE user_id = $1 AND consumed_at IS NULL`, userID)
	return err
}
