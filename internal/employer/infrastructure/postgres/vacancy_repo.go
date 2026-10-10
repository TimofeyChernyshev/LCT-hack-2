package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type VacancyRepo struct{ pool *pgxpool.Pool }

func NewVacancyRepo(pool *pgxpool.Pool) *VacancyRepo { return &VacancyRepo{pool: pool} }

func (r *VacancyRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

const vacancyColumns = `id, company_id, need_id, title, description,
    category_id, specialization_id, grade_id, stack::text[],
    salary_min, salary_max, currency, work_format, location,
    status, published_at, created_at, updated_at`

func (r *VacancyRepo) scan(row pgx.Row, v *domain.Vacancy) error {
	var wf *string
	var status string
	err := row.Scan(
		&v.ID, &v.CompanyID, &v.NeedID, &v.Title, &v.Description,
		&v.CategoryID, &v.SpecializationID, &v.GradeID, &v.Stack,
		&v.SalaryMin, &v.SalaryMax, &v.Currency, &wf, &v.Location,
		&status, &v.PublishedAt, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if wf != nil {
		f := domain.WorkFormat(*wf)
		v.WorkFormat = &f
	}
	v.Status = domain.VacancyStatus(status)
	return nil
}

func (r *VacancyRepo) ListByCompany(ctx context.Context, companyID string) ([]domain.Vacancy, error) {
	rows, err := r.q(ctx).Query(ctx,
		`SELECT `+vacancyColumns+` FROM vacancies WHERE company_id = $1 ORDER BY created_at DESC`,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Vacancy
	for rows.Next() {
		var v domain.Vacancy
		if err := r.scan(rows, &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *VacancyRepo) ListPublic(ctx context.Context, f application.PublicVacancyFilter) ([]domain.Vacancy, int, error) {
	where := []string{"status = 'published'"}
	args := []any{}
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		where = append(where, "category_id = $"+strconv.Itoa(len(args)))
	}
	if f.GradeID != nil {
		args = append(args, *f.GradeID)
		where = append(where, "grade_id = $"+strconv.Itoa(len(args)))
	}

	whereClause := " WHERE " + strings.Join(where, " AND ")

	var total int
	if err := r.q(ctx).QueryRow(ctx,
		`SELECT COUNT(*) FROM vacancies`+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, f.Limit, f.Offset)
	q := `SELECT ` + vacancyColumns + ` FROM vacancies` + whereClause +
		` ORDER BY published_at DESC NULLS LAST LIMIT $` + strconv.Itoa(len(listArgs)-1) +
		` OFFSET $` + strconv.Itoa(len(listArgs))

	rows, err := r.q(ctx).Query(ctx, q, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Vacancy
	for rows.Next() {
		var v domain.Vacancy
		if err := r.scan(rows, &v); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func (r *VacancyRepo) GetByID(ctx context.Context, id string) (*domain.Vacancy, error) {
	var v domain.Vacancy
	err := r.scan(r.q(ctx).QueryRow(ctx,
		`SELECT `+vacancyColumns+` FROM vacancies WHERE id = $1`, id), &v)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrVacancyNotFound
	}
	return &v, err
}

func (r *VacancyRepo) Create(ctx context.Context, v *domain.Vacancy) error {
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO vacancies
			(company_id, need_id, title, description,
			 category_id, specialization_id, grade_id, stack,
			 salary_min, salary_max, currency, work_format, location, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::uuid[],$9,$10,$11,$12,$13,$14)
		RETURNING id, created_at, updated_at`,
		v.CompanyID, v.NeedID, v.Title, v.Description,
		v.CategoryID, v.SpecializationID, v.GradeID, v.Stack,
		v.SalaryMin, v.SalaryMax, v.Currency, v.WorkFormat, v.Location, string(v.Status),
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func (r *VacancyRepo) Update(ctx context.Context, v *domain.Vacancy) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE vacancies SET
			need_id = $2, title = $3, description = $4,
			category_id = $5, specialization_id = $6, grade_id = $7, stack = $8::uuid[],
			salary_min = $9, salary_max = $10, currency = $11,
			work_format = $12, location = $13, updated_at = now()
		WHERE id = $1`,
		v.ID, v.NeedID, v.Title, v.Description,
		v.CategoryID, v.SpecializationID, v.GradeID, v.Stack,
		v.SalaryMin, v.SalaryMax, v.Currency, v.WorkFormat, v.Location)
	return err
}

func (r *VacancyRepo) Delete(ctx context.Context, id, companyID string) error {
	tag, err := r.q(ctx).Exec(ctx,
		`DELETE FROM vacancies WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVacancyNotFound
	}
	return nil
}

func (r *VacancyRepo) Publish(ctx context.Context, id, companyID string) error {
	tag, err := r.q(ctx).Exec(ctx, `
		UPDATE vacancies
		SET status = 'published', published_at = now(), updated_at = now()
		WHERE id = $1 AND company_id = $2 AND status <> 'archived'`, id, companyID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVacancyNotFound
	}
	return nil
}
