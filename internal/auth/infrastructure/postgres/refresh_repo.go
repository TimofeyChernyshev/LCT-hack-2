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

type RefreshTokenRepo struct{ pool *pgxpool.Pool }

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

func (r *RefreshTokenRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *RefreshTokenRepo) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time, ip, ua string) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip, user_agent)
		VALUES ($1, $2, $3, NULLIF($4,'')::inet, $5)`,
		userID, tokenHash, expiresAt, ip, ua)
	return err
}

func (r *RefreshTokenRepo) GetForUpdate(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	var revokedAt *time.Time
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, user_id, revoked_at, expires_at
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE`, tokenHash).
		Scan(&t.ID, &t.UserID, &revokedAt, &t.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	t.Revoked = revokedAt != nil
	return &t, nil
}

func (r *RefreshTokenRepo) MarkRevoked(ctx context.Context, id string) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1`, id)
	return err
}

func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}
