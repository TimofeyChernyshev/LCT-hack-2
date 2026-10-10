package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type VisibilityRepo struct{ pool *pgxpool.Pool }

func NewVisibilityRepo(pool *pgxpool.Pool) *VisibilityRepo { return &VisibilityRepo{pool: pool} }

func (r *VisibilityRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *VisibilityRepo) Get(ctx context.Context, userID string) (*domain.Visibility, error) {
	var v domain.Visibility
	err := r.q(ctx).QueryRow(ctx, `
		SELECT contacts, links, fsp, experience, resume, salary, soft_skills
		FROM candidate_visibility WHERE user_id = $1`, userID,
	).Scan(&v.Contacts, &v.Links, &v.FSP, &v.Experience, &v.Resume, &v.Salary, &v.SoftSkills)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VisibilityRepo) Upsert(ctx context.Context, userID string, v *domain.Visibility) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO candidate_visibility
			(user_id, contacts, links, fsp, experience, resume, salary, soft_skills, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (user_id) DO UPDATE SET
			contacts    = EXCLUDED.contacts,
			links       = EXCLUDED.links,
			fsp         = EXCLUDED.fsp,
			experience  = EXCLUDED.experience,
			resume      = EXCLUDED.resume,
			salary      = EXCLUDED.salary,
			soft_skills = EXCLUDED.soft_skills,
			updated_at  = now()`,
		userID, v.Contacts, v.Links, v.FSP, v.Experience, v.Resume, v.Salary, v.SoftSkills)
	return err
}
