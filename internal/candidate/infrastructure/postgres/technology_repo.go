package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type TechnologyRepo struct{ pool *pgxpool.Pool }

func NewTechnologyRepo(pool *pgxpool.Pool) *TechnologyRepo { return &TechnologyRepo{pool: pool} }

func (r *TechnologyRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *TechnologyRepo) ListByUser(ctx context.Context, userID string) ([]domain.CandidateTechnology, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT user_id, technology_id, level
		FROM candidate_technologies WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.CandidateTechnology
	for rows.Next() {
		var t domain.CandidateTechnology
		if err := rows.Scan(&t.UserID, &t.TechnologyID, &t.Level); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TechnologyRepo) ReplaceForUser(ctx context.Context, userID string, techs []domain.CandidateTechnology) error {
	if _, err := r.q(ctx).Exec(ctx,
		`DELETE FROM candidate_technologies WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, t := range techs {
		if _, err := r.q(ctx).Exec(ctx, `
			INSERT INTO candidate_technologies (user_id, technology_id, level)
			VALUES ($1, $2, $3)`, userID, t.TechnologyID, t.Level); err != nil {
			return err
		}
	}
	return nil
}
