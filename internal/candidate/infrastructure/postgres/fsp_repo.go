package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/domain"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

type FSPRepo struct{ pool *pgxpool.Pool }

func NewFSPRepo(pool *pgxpool.Pool) *FSPRepo { return &FSPRepo{pool: pool} }

func (r *FSPRepo) q(ctx context.Context) pkgpostgres.Queryable {
	return pkgpostgres.QueryableFromContext(ctx, r.pool)
}

func (r *FSPRepo) GetMemberID(ctx context.Context, userID string) (*string, *time.Time, error) {
	var id *string
	var at *time.Time
	err := r.q(ctx).QueryRow(ctx, `
		SELECT fsp_member_id, fsp_linked_at FROM candidate_profiles WHERE user_id = $1`, userID,
	).Scan(&id, &at)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, domain.ErrProfileNotFound
	}
	return id, at, err
}

func (r *FSPRepo) LinkMemberID(ctx context.Context, userID, memberID string) error {
	_, err := r.q(ctx).Exec(ctx, `
		UPDATE candidate_profiles
		SET fsp_member_id = $2, fsp_linked_at = now(), updated_at = now()
		WHERE user_id = $1`, userID, memberID)
	return err
}

func (r *FSPRepo) ListAchievements(ctx context.Context, userID string) ([]domain.FSPAchievement, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, user_id, external_id, event_name, event_date, place, category, score, weight, source
		FROM fsp_achievements WHERE user_id = $1
		ORDER BY event_date DESC NULLS LAST, weight DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.FSPAchievement
	for rows.Next() {
		var a domain.FSPAchievement
		if err := rows.Scan(&a.ID, &a.UserID, &a.ExternalID, &a.EventName, &a.EventDate,
			&a.Place, &a.Category, &a.Score, &a.Weight, &a.Source); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *FSPRepo) CountAchievements(ctx context.Context, userID string) (int, *int, error) {
	var count int
	var best *int
	err := r.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*), MIN(place) FROM fsp_achievements WHERE user_id = $1`, userID,
	).Scan(&count, &best)
	return count, best, err
}

func (r *FSPRepo) CreateAchievement(ctx context.Context, a *domain.FSPAchievement) error {
	return r.q(ctx).QueryRow(ctx, `
		INSERT INTO fsp_achievements (user_id, external_id, event_name, event_date, place, category, score, weight, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`,
		a.UserID, a.ExternalID, a.EventName, a.EventDate, a.Place,
		a.Category, a.Score, a.Weight, a.Source,
	).Scan(&a.ID)
}
