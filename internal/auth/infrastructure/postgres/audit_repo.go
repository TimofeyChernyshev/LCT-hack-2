package postgres

import (
	"context"
	"encoding/json"

	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct{ pool *pgxpool.Pool }

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo { return &AuditRepo{pool: pool} }

func (r *AuditRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *AuditRepo) Log(ctx context.Context, userID, action, ip, ua string, payload map[string]any) error {
	body := []byte("{}")
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			body = b
		}
	}
	var uid *string
	if userID != "" {
		uid = &userID
	}
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO audit_log (user_id, action, ip, user_agent, payload)
		VALUES ($1, $2, NULLIF($3,'')::inet, $4, $5)`,
		uid, action, ip, ua, body)
	return err
}
