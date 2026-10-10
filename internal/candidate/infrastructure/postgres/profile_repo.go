package postgres

import (
	"context"
	"errors"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepo struct{ pool *pgxpool.Pool }

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo { return &ProfileRepo{pool: pool} }

func (r *ProfileRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

const profileColumns = `user_id, first_name, last_name, middle_name,
    headline, about, location, years_experience,
    current_category_id, current_grade_id, specialization_id, fsp_member_id,
    salary_min, salary_max, salary_currency, soft_skills, updated_at`

func (r *ProfileRepo) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	var p domain.Profile
	err := r.q(ctx).QueryRow(ctx,
		`SELECT `+profileColumns+` FROM candidate_profiles WHERE user_id = $1`, userID,
	).Scan(
		&p.UserID, &p.FirstName, &p.LastName, &p.MiddleName,
		&p.Headline, &p.About, &p.Location, &p.YearsExperience,
		&p.CategoryID, &p.GradeID, &p.SpecializationID, &p.FSPMemberID,
		&p.SalaryMin, &p.SalaryMax, &p.SalaryCurrency, &p.SoftSkills, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrProfileNotFound
	}
	return &p, err
}

func skillsOrEmpty(skills []string) []string {
	if skills == nil {
		return []string{}
	}
	return skills
}

func (r *ProfileRepo) Create(ctx context.Context, p *domain.Profile) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO candidate_profiles
			(user_id, first_name, last_name, middle_name,
			 headline, about, location, years_experience,
			 salary_min, salary_max, salary_currency, soft_skills, specialization_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		p.UserID, p.FirstName, p.LastName, p.MiddleName,
		p.Headline, p.About, p.Location, p.YearsExperience,
		p.SalaryMin, p.SalaryMax, p.SalaryCurrency, skillsOrEmpty(p.SoftSkills), p.SpecializationID)
	return err
}

func (r *ProfileRepo) Update(ctx context.Context, p *domain.Profile) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE candidate_profiles SET
			first_name = $2, last_name = $3, middle_name = $4,
			headline = $5, about = $6, location = $7, years_experience = $8,
			salary_min = $9, salary_max = $10, salary_currency = $11,
			soft_skills = $12, specialization_id = $13, updated_at = now()
		WHERE user_id = $1`,
		p.UserID, p.FirstName, p.LastName, p.MiddleName,
		p.Headline, p.About, p.Location, p.YearsExperience,
		p.SalaryMin, p.SalaryMax, p.SalaryCurrency, skillsOrEmpty(p.SoftSkills), p.SpecializationID)
	return err
}
