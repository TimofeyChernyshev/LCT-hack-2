package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type ContactsRepo struct{ pool *pgxpool.Pool }

func NewContactsRepo(pool *pgxpool.Pool) *ContactsRepo { return &ContactsRepo{pool: pool} }

func (r *ContactsRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *ContactsRepo) GetByUserID(ctx context.Context, userID string) (*domain.Contacts, error) {
	var c domain.Contacts
	var email *string
	err := r.q(ctx).QueryRow(ctx, `
		SELECT email, phone, telegram, github, linkedin, website
		FROM candidate_contacts WHERE user_id = $1`, userID,
	).Scan(&email, &c.Phone, &c.Telegram, &c.GitHub, &c.LinkedIn, &c.Website)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrContactsNotFound
	}
	if err != nil {
		return nil, err
	}
	if email != nil {
		c.Email = *email
	}
	return &c, nil
}

func (r *ContactsRepo) Upsert(ctx context.Context, userID string, c *domain.Contacts) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO candidate_contacts
			(user_id, email, phone, telegram, github, linkedin, website, updated_at)
		VALUES ($1, NULLIF($2,''), $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id) DO UPDATE SET
			email      = EXCLUDED.email,
			phone      = EXCLUDED.phone,
			telegram   = EXCLUDED.telegram,
			github     = EXCLUDED.github,
			linkedin   = EXCLUDED.linkedin,
			website    = EXCLUDED.website,
			updated_at = now()`,
		userID, c.Email, c.Phone, c.Telegram, c.GitHub, c.LinkedIn, c.Website)
	return err
}
