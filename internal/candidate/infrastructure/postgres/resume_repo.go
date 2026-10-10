package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type ResumeRepo struct{ pool *pgxpool.Pool }

func NewResumeRepo(pool *pgxpool.Pool) *ResumeRepo { return &ResumeRepo{pool: pool} }

func (r *ResumeRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *ResumeRepo) ListByUser(ctx context.Context, userID string) ([]domain.Resume, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, user_id, title, source, content, file_path, parsed, is_primary, created_at, updated_at
		FROM resumes WHERE user_id = $1 ORDER BY is_primary DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Resume
	for rows.Next() {
		var x domain.Resume
		var src string
		if err := rows.Scan(&x.ID, &x.UserID, &x.Title, &src, &x.Content, &x.FilePath,
			&x.Parsed, &x.IsPrimary, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		x.Source = domain.ResumeSource(src)
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *ResumeRepo) GetByID(ctx context.Context, id string) (*domain.Resume, error) {
	var x domain.Resume
	var src string
	err := r.q(ctx).QueryRow(ctx, `
		SELECT id, user_id, title, source, content, file_path, parsed, is_primary, created_at, updated_at
		FROM resumes WHERE id = $1`, id,
	).Scan(&x.ID, &x.UserID, &x.Title, &src, &x.Content, &x.FilePath,
		&x.Parsed, &x.IsPrimary, &x.CreatedAt, &x.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrResumeNotFound
	}
	if err != nil {
		return nil, err
	}
	x.Source = domain.ResumeSource(src)
	return &x, nil
}

func (r *ResumeRepo) Create(ctx context.Context, x *domain.Resume) error {
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO resumes (user_id, title, source, content, file_path, parsed, is_primary)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at, updated_at`,
		x.UserID, x.Title, string(x.Source), x.Content, x.FilePath, x.Parsed, x.IsPrimary,
	).Scan(&x.ID, &x.CreatedAt, &x.UpdatedAt)
}

func (r *ResumeRepo) UnsetPrimary(ctx context.Context, userID string) error {
	_, err := r.q(ctx).Exec(ctx,
		`UPDATE resumes SET is_primary = false, updated_at = now() WHERE user_id = $1`, userID)
	return err
}
