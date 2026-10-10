package postgres

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsentRepo struct{ pool *pgxpool.Pool }

func NewConsentRepo(pool *pgxpool.Pool) *ConsentRepo { return &ConsentRepo{pool: pool} }

func (r *ConsentRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

// Grant — идемпотентно за счёт частичного уникального индекса
// idx_user_consents_active (user_id, type) WHERE revoked_at IS NULL.
// Бизнес-логика «не дублировать» — в БД через constraint, а не в коде.
func (r *ConsentRepo) Grant(ctx context.Context, c *domain.Consent) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO user_consents (user_id, type, version, ip, user_agent)
		VALUES ($1, $2, $3, NULLIF($4,'')::inet, $5)
		ON CONFLICT (user_id, type) WHERE revoked_at IS NULL DO NOTHING`,
		c.UserID, string(c.Type), c.Version, c.IP, c.UserAgent)
	return err
}

func (r *ConsentRepo) Revoke(ctx context.Context, userID string, t domain.ConsentType) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE user_consents
		SET revoked_at = now()
		WHERE user_id = $1 AND type = $2 AND revoked_at IS NULL`,
		userID, string(t))
	return err
}

func (r *ConsentRepo) ListByUser(ctx context.Context, userID string) ([]domain.Consent, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, user_id, type, version, granted_at, revoked_at,
		       COALESCE(host(ip), ''), COALESCE(user_agent, '')
		FROM user_consents
		WHERE user_id = $1
		ORDER BY granted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Consent
	for rows.Next() {
		var c domain.Consent
		var t string
		if err := rows.Scan(&c.ID, &c.UserID, &t, &c.Version,
			&c.GrantedAt, &c.RevokedAt, &c.IP, &c.UserAgent); err != nil {
			return nil, err
		}
		c.Type = domain.ConsentType(t)
		out = append(out, c)
	}
	return out, rows.Err()
}
