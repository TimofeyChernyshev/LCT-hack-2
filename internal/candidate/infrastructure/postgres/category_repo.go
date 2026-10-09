package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type CategoryRepo struct{ pool *pgxpool.Pool }

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo { return &CategoryRepo{pool: pool} }

func (r *CategoryRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *CategoryRepo) GetState(ctx context.Context, userID string) (*domain.CategoryState, error) {
	st := &domain.CategoryState{}

	err := r.q(ctx).QueryRow(ctx, `
		SELECT current_category_id, current_grade_id, specialization_id
		FROM candidate_profiles WHERE user_id = $1`, userID,
	).Scan(&st.CategoryID, &st.GradeID, &st.SpecializationID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	rows, err := r.q(ctx).Query(ctx, `
		SELECT category_id, grade_id, specialization_id, reason, effective_from, effective_to
		FROM candidate_category_history
		WHERE user_id = $1
		ORDER BY effective_from DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h domain.CategoryHistoryEntry
		if err := rows.Scan(&h.CategoryID, &h.GradeID, &h.SpecializationID,
			&h.Reason, &h.EffectiveFrom, &h.EffectiveTo); err != nil {
			return nil, err
		}
		st.History = append(st.History, h)
	}
	return st, rows.Err()
}
