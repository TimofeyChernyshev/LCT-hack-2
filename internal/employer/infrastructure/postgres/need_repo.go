package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type NeedRepo struct{ pool *pgxpool.Pool }

func NewNeedRepo(pool *pgxpool.Pool) *NeedRepo { return &NeedRepo{pool: pool} }

func (r *NeedRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

const needColumns = `id, company_id, title, description,
    category_id, specialization_id, grade_id, stack::text[],
    salary_min, salary_max, currency, work_format,
    location, min_experience_years, status, created_at, updated_at`

func (r *NeedRepo) scan(row pgx.Row, n *domain.Need) error {
	var wf *string
	var status string
	err := row.Scan(
		&n.ID, &n.CompanyID, &n.Title, &n.Description,
		&n.CategoryID, &n.SpecializationID, &n.GradeID, &n.Stack,
		&n.SalaryMin, &n.SalaryMax, &n.Currency, &wf,
		&n.Location, &n.MinExperienceYears, &status, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if wf != nil {
		f := domain.WorkFormat(*wf)
		n.WorkFormat = &f
	}
	n.Status = domain.NeedStatus(status)
	return nil
}

func (r *NeedRepo) ListByCompany(ctx context.Context, companyID string) ([]domain.Need, error) {
	rows, err := r.q(ctx).Query(ctx,
		`SELECT `+needColumns+` FROM employer_needs WHERE company_id = $1 ORDER BY created_at DESC`,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Need
	for rows.Next() {
		var n domain.Need
		if err := r.scan(rows, &n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *NeedRepo) GetByID(ctx context.Context, id string) (*domain.Need, error) {
	var n domain.Need
	err := r.scan(r.q(ctx).QueryRow(ctx,
		`SELECT `+needColumns+` FROM employer_needs WHERE id = $1`, id), &n)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNeedNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NeedRepo) Create(ctx context.Context, n *domain.Need) error {
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO employer_needs
			(company_id, title, description, category_id, specialization_id, grade_id,
			 stack, salary_min, salary_max, currency, work_format, location,
			 min_experience_years, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7::uuid[],$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, updated_at`,
		n.CompanyID, n.Title, n.Description, n.CategoryID, n.SpecializationID, n.GradeID,
		n.Stack, n.SalaryMin, n.SalaryMax, n.Currency, n.WorkFormat, n.Location,
		n.MinExperienceYears, string(n.Status),
	).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
}

func (r *NeedRepo) Update(ctx context.Context, n *domain.Need) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE employer_needs SET
			title = $2, description = $3,
			category_id = $4, specialization_id = $5, grade_id = $6,
			stack = $7::uuid[],
			salary_min = $8, salary_max = $9, currency = $10,
			work_format = $11, location = $12, min_experience_years = $13,
			status = $14, updated_at = now()
		WHERE id = $1`,
		n.ID, n.Title, n.Description, n.CategoryID, n.SpecializationID, n.GradeID,
		n.Stack, n.SalaryMin, n.SalaryMax, n.Currency, n.WorkFormat, n.Location,
		n.MinExperienceYears, string(n.Status))
	return err
}

func (r *NeedRepo) Delete(ctx context.Context, id, companyID string) error {
	tag, err := r.q(ctx).Exec(ctx,
		`DELETE FROM employer_needs WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNeedNotFound
	}
	return nil
}
