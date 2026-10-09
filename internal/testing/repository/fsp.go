package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	"TimofeyChernyshev/LCT-hack-2/pkg/fsp"
)

func (r *PostgresRepository) SaveCandidateFSP(ctx context.Context, profile *domain.CandidateFSPProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	profileQuery := `
		INSERT INTO candidate_fsp_profiles (
			user_id, fsp_member_id, full_name, sports_rank, fsp_rating,
			region, discipline, has_fsp, fsp_score, fsp_weight_sum,
			achievements_count, best_place, verification_source, linked_at,
			explanation, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (user_id) DO UPDATE SET
			fsp_member_id       = EXCLUDED.fsp_member_id,
			full_name           = EXCLUDED.full_name,
			sports_rank         = EXCLUDED.sports_rank,
			fsp_rating          = EXCLUDED.fsp_rating,
			region              = EXCLUDED.region,
			discipline          = EXCLUDED.discipline,
			has_fsp             = EXCLUDED.has_fsp,
			fsp_score           = EXCLUDED.fsp_score,
			fsp_weight_sum      = EXCLUDED.fsp_weight_sum,
			achievements_count  = EXCLUDED.achievements_count,
			best_place          = EXCLUDED.best_place,
			verification_source = EXCLUDED.verification_source,
			linked_at           = EXCLUDED.linked_at,
			explanation         = EXCLUDED.explanation,
			updated_at          = EXCLUDED.updated_at;
	`

	now := time.Now()
	_, err = tx.ExecContext(ctx, profileQuery,
		profile.UserID, profile.FSPMemberID, profile.FullName, profile.SportsRank, profile.FSPRating,
		profile.Region, profile.Discipline, profile.HasFSP, profile.FSPScore, profile.FSPWeightSum,
		profile.AchievementsCount, profile.BestPlace, profile.VerificationSource, profile.LinkedAt,
		profile.Explanation, now,
	)
	if err != nil {
		return fmt.Errorf("upsert candidate fsp profile: %w", err)
	}

	// Delete old achievements
	_, err = tx.ExecContext(ctx, "DELETE FROM candidate_fsp_achievements WHERE user_id = $1;", profile.UserID)
	if err != nil {
		return fmt.Errorf("delete old fsp achievements: %w", err)
	}

	// Insert new achievements
	if len(profile.Achievements) > 0 {
		insertAchQuery := `
			INSERT INTO candidate_fsp_achievements (
				id, user_id, external_id, event_name, event_date, place,
				category, score, weight, badge, description, payload, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (user_id, external_id) DO UPDATE SET
				event_name  = EXCLUDED.event_name,
				event_date  = EXCLUDED.event_date,
				place       = EXCLUDED.place,
				category    = EXCLUDED.category,
				score       = EXCLUDED.score,
				weight      = EXCLUDED.weight,
				badge       = EXCLUDED.badge,
				description = EXCLUDED.description,
				payload     = EXCLUDED.payload;
		`

		for _, ach := range profile.Achievements {
			achID := uuid.New()
			if parsed, err := uuid.Parse(ach.ID); err == nil {
				achID = parsed
			}

			var eventDate sql.NullTime
			if ach.EventDate != "" {
				if t, err := time.Parse("2006-01-02", ach.EventDate); err == nil {
					eventDate = sql.NullTime{Time: t, Valid: true}
				}
			}

			payloadBytes, _ := json.Marshal(ach.Payload)
			if payloadBytes == nil {
				payloadBytes = []byte("{}")
			}

			_, err = tx.ExecContext(ctx, insertAchQuery,
				achID, profile.UserID, ach.ExternalID, ach.EventName, eventDate, ach.Place,
				ach.Category, ach.Score, ach.Weight, ach.Badge, ach.Description, payloadBytes, now,
			)
			if err != nil {
				return fmt.Errorf("insert fsp achievement: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetCandidateFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error) {
	profileQuery := `
		SELECT user_id, fsp_member_id, full_name, sports_rank, fsp_rating,
		       region, discipline, has_fsp, fsp_score, fsp_weight_sum,
		       achievements_count, best_place, verification_source, linked_at,
		       explanation, created_at, updated_at
		FROM candidate_fsp_profiles
		WHERE user_id = $1;
	`

	var p domain.CandidateFSPProfile
	var fspMemberID, fullName, sportsRank, region, discipline sql.NullString
	var bestPlace sql.NullInt32
	var linkedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, profileQuery, userID).Scan(
		&p.UserID, &fspMemberID, &fullName, &sportsRank, &p.FSPRating,
		&region, &discipline, &p.HasFSP, &p.FSPScore, &p.FSPWeightSum,
		&p.AchievementsCount, &bestPlace, &p.VerificationSource, &linkedAt,
		&p.Explanation, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query candidate fsp profile: %w", err)
	}

	if fspMemberID.Valid {
		p.FSPMemberID = &fspMemberID.String
	}
	if fullName.Valid {
		p.FullName = &fullName.String
	}
	if sportsRank.Valid {
		p.SportsRank = &sportsRank.String
	}
	if region.Valid {
		p.Region = &region.String
	}
	if discipline.Valid {
		p.Discipline = &discipline.String
	}
	if bestPlace.Valid {
		bp := int(bestPlace.Int32)
		p.BestPlace = &bp
	}
	if linkedAt.Valid {
		p.LinkedAt = &linkedAt.Time
	}

	// Query achievements
	achQuery := `
		SELECT id, external_id, event_name, event_date, place,
		       category, score, weight, badge, description, payload
		FROM candidate_fsp_achievements
		WHERE user_id = $1
		ORDER BY weight DESC, place ASC NULLS LAST;
	`
	rows, err := r.db.QueryContext(ctx, achQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("query fsp achievements: %w", err)
	}
	defer rows.Close()

	var achievements []fsp.Achievement
	for rows.Next() {
		var ach fsp.Achievement
		var achID uuid.UUID
		var extID, badge, desc sql.NullString
		var eventDate sql.NullTime
		var place sql.NullInt32
		var score sql.NullFloat64
		var payloadBytes []byte

		if err := rows.Scan(
			&achID, &extID, &ach.EventName, &eventDate, &place,
			&ach.Category, &score, &ach.Weight, &badge, &desc, &payloadBytes,
		); err != nil {
			return nil, fmt.Errorf("scan fsp achievement: %w", err)
		}

		ach.ID = achID.String()
		if extID.Valid {
			ach.ExternalID = extID.String
		}
		if eventDate.Valid {
			ach.EventDate = eventDate.Time.Format("2006-01-02")
		}
		if place.Valid {
			pl := int(place.Int32)
			ach.Place = &pl
		}
		if score.Valid {
			sc := score.Float64
			ach.Score = &sc
		}
		if badge.Valid {
			ach.Badge = badge.String
		}
		if desc.Valid {
			ach.Description = desc.String
		}
		if len(payloadBytes) > 0 {
			_ = json.Unmarshal(payloadBytes, &ach.Payload)
		}

		achievements = append(achievements, ach)
	}

	p.Achievements = achievements
	return &p, nil
}

func (r *PostgresRepository) UnlinkCandidateFSP(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete achievements
	_, err = tx.ExecContext(ctx, "DELETE FROM candidate_fsp_achievements WHERE user_id = $1;", userID)
	if err != nil {
		return fmt.Errorf("delete fsp achievements: %w", err)
	}

	// Reset profile to clean neutral state
	resetQuery := `
		UPDATE candidate_fsp_profiles SET
			has_fsp             = FALSE,
			fsp_member_id       = NULL,
			full_name           = NULL,
			sports_rank         = NULL,
			fsp_rating          = 0,
			region              = NULL,
			discipline          = NULL,
			fsp_score           = 0.00,
			fsp_weight_sum      = 0,
			achievements_count  = 0,
			best_place          = NULL,
			linked_at           = NULL,
			explanation         = 'История участия в соревнованиях ФСП не привязана. Кандидат оценивается по результатам тестов и стеку компетенций.',
			updated_at          = now()
		WHERE user_id = $1;
	`
	_, err = tx.ExecContext(ctx, resetQuery, userID)
	if err != nil {
		return fmt.Errorf("reset candidate fsp profile: %w", err)
	}

	return tx.Commit()
}
