package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type CompanyRepo struct{ pool *pgxpool.Pool }

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo { return &CompanyRepo{pool: pool} }

func (r *CompanyRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

const companyColumns = `id, owner_user_id, name, description, industry, website, size,
    contact_person, contact_email, contact_phone, contact_telegram, created_at, updated_at`

func (r *CompanyRepo) GetByOwner(ctx context.Context, ownerID string) (*domain.Company, error) {
	var c domain.Company
	err := r.q(ctx).QueryRow(ctx,
		`SELECT `+companyColumns+` FROM companies WHERE owner_user_id = $1`, ownerID,
	).Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.Description, &c.Industry, &c.Website,
		&c.Size, &c.ContactPerson, &c.ContactEmail, &c.ContactPhone, &c.ContactTelegram,
		&c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCompanyNotFound
	}
	return &c, err
}

func (r *CompanyRepo) Create(ctx context.Context, c *domain.Company) error {
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO companies
			(owner_user_id, name, description, industry, website, size,
			 contact_person, contact_email, contact_phone, contact_telegram)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`,
		c.OwnerUserID, c.Name, c.Description, c.Industry, c.Website, c.Size,
		c.ContactPerson, c.ContactEmail, c.ContactPhone, c.ContactTelegram,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CompanyRepo) Update(ctx context.Context, c *domain.Company) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE companies SET
			name = $2, description = $3, industry = $4, website = $5, size = $6,
			contact_person = $7, contact_email = $8, contact_phone = $9, contact_telegram = $10,
			updated_at = now()
		WHERE id = $1`,
		c.ID, c.Name, c.Description, c.Industry, c.Website, c.Size,
		c.ContactPerson, c.ContactEmail, c.ContactPhone, c.ContactTelegram)
	return err
}
