package postgres

import (
	"context"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RevealRepo struct{ pool *pgxpool.Pool }

func NewRevealRepo(pool *pgxpool.Pool) *RevealRepo { return &RevealRepo{pool: pool} }

func (r *RevealRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *RevealRepo) IsRevealed(ctx context.Context, candidateUserID, employerUserID string) (bool, error) {
	var exists bool
	err := r.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM contact_reveals
			WHERE candidate_user_id = $1 AND employer_user_id = $2
		)`, candidateUserID, employerUserID).Scan(&exists)
	return exists, err
}

func (r *RevealRepo) Create(ctx context.Context, rv *domain.ContactReveal) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO contact_reveals
			(candidate_user_id, employer_user_id, entity_type, entity_id, reason)
		VALUES ($1, $2, $3, $4, $5)`,
		rv.CandidateUserID, rv.EmployerUserID,
		string(rv.EntityType), rv.EntityID, string(rv.Reason))
	return err
}
