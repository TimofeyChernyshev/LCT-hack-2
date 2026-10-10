package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type ExperienceRepo struct{ pool *pgxpool.Pool }

func NewExperienceRepo(pool *pgxpool.Pool) *ExperienceRepo { return &ExperienceRepo{pool: pool} }

func (r *ExperienceRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *ExperienceRepo) ListByUser(ctx context.Context, userID string) ([]domain.Experience, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, user_id, company, position, started_at, ended_at, description, created_at
		FROM candidate_experiences WHERE user_id = $1 ORDER BY started_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Experience
	for rows.Next() {
		var e domain.Experience
		var ended *time.Time
		if err := rows.Scan(&e.ID, &e.UserID, &e.Company, &e.Position,
			&e.StartedAt, &ended, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		if ended != nil {
			e.EndedAt = *ended
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *ExperienceRepo) Create(ctx context.Context, e *domain.Experience) error {
	var ended any
	if !e.EndedAt.IsZero() {
		ended = e.EndedAt
	}
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO candidate_experiences (user_id, company, position, started_at, ended_at, description)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at`,
		e.UserID, e.Company, e.Position, e.StartedAt, ended, e.Description,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *ExperienceRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.q(ctx).Exec(ctx, `
		DELETE FROM candidate_experiences WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrExperienceNotFound
	}
	return nil
}
