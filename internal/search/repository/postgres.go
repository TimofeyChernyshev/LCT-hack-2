package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNotFound = errors.New("candidate search document not found")
)

type Repository interface {
	Search(ctx context.Context, filter domain.SearchFilter) ([]domain.CandidateSearchDoc, int, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.CandidateSearchDoc, error)
	GetCategoryRanking(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.CandidateSearchDoc, error)
	CalculateTestPercentile(ctx context.Context, categoryID uuid.UUID, testScore float64) (float64, error)
	Upsert(ctx context.Context, doc *domain.CandidateSearchDoc) error
	Count(ctx context.Context) (int, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Search(ctx context.Context, filter domain.SearchFilter) ([]domain.CandidateSearchDoc, int, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	if filter.CategoryID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.SpecializationID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("specialization_id = $%d", argIdx))
		args = append(args, *filter.SpecializationID)
		argIdx++
	}

	if filter.GradeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("grade_id = $%d", argIdx))
		args = append(args, *filter.GradeID)
		argIdx++
	}

	if filter.HasFSP != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("has_fsp = $%d", argIdx))
		args = append(args, *filter.HasFSP)
		argIdx++
	}

	if filter.MinYearsExperience != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("years_experience >= $%d", argIdx))
		args = append(args, *filter.MinYearsExperience)
		argIdx++
	}

	if filter.Location != nil && strings.TrimSpace(*filter.Location) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("location ILIKE $%d", argIdx))
		args = append(args, "%"+strings.TrimSpace(*filter.Location)+"%")
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sorting
	orderBySQL := "ORDER BY calculated_score DESC, test_score DESC NULLS LAST"
	switch filter.Sort {
	case "testScore":
		orderBySQL = "ORDER BY test_score DESC NULLS LAST, calculated_score DESC"
	case "fspWeight":
		orderBySQL = "ORDER BY fsp_weight_sum DESC, fsp_score DESC, calculated_score DESC"
	case "relevance":
		fallthrough
	default:
		orderBySQL = "ORDER BY calculated_score DESC, test_score DESC NULLS LAST"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT user_id, display_name, category_id, specialization_id, specialization_name,
		       grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
		       fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
		       fsp_highlights::text, stack::text, years_experience, location, periodic_tasks_solved,
		       activity_score, calculated_score, explanation, reasons::text, last_active_at, updated_at,
		       COUNT(*) OVER() AS total_count
		FROM candidate_search_docs
		%s
		%s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, orderBySQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var docs []domain.CandidateSearchDoc
	total := 0

	for rows.Next() {
		var d domain.CandidateSearchDoc
		var testScore sql.NullFloat64
		var bestPlace sql.NullInt32
		var yearsExp sql.NullFloat64
		var location sql.NullString
		var sportsRank sql.NullString
		var fspHighlightsStr, stackStr, reasonsStr string
		var count int

		err := rows.Scan(
			&d.UserID, &d.DisplayName, &d.CategoryID, &d.SpecializationID, &d.SpecializationName,
			&d.GradeID, &d.GradeName, &d.GradeRank, &testScore, &d.HasFSP, &sportsRank,
			&d.FSPRating, &d.FSPScore, &d.FSPAchievementsCount, &bestPlace, &d.FSPWeightSum,
			&fspHighlightsStr, &stackStr, &yearsExp, &location, &d.PeriodicTasksSolved,
			&d.ActivityScore, &d.CalculatedScore, &d.Explanation, &reasonsStr, &d.LastActiveAt, &d.UpdatedAt,
			&count,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan candidate doc: %w", err)
		}

		total = count
		if testScore.Valid {
			d.TestScore = &testScore.Float64
		}
		if bestPlace.Valid {
			bp := int(bestPlace.Int32)
			d.FSPBestPlace = &bp
		}
		if yearsExp.Valid {
			d.YearsExperience = &yearsExp.Float64
		}
		if location.Valid {
			d.Location = &location.String
		}
		if sportsRank.Valid {
			d.SportsRank = sportsRank.String
		}

		d.FSPHighlights = parseStringArray(fspHighlightsStr)
		d.Stack = parseUUIDArray(stackStr)
		d.Reasons = parseStringArray(reasonsStr)

		docs = append(docs, d)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	return docs, total, nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.CandidateSearchDoc, error) {
	query := `
		SELECT user_id, display_name, category_id, specialization_id, specialization_name,
		       grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
		       fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
		       fsp_highlights::text, stack::text, years_experience, location, periodic_tasks_solved,
		       activity_score, calculated_score, explanation, reasons::text, last_active_at, updated_at
		FROM candidate_search_docs
		WHERE user_id = $1;
	`

	var d domain.CandidateSearchDoc
	var testScore sql.NullFloat64
	var bestPlace sql.NullInt32
	var yearsExp sql.NullFloat64
	var location sql.NullString
	var sportsRank sql.NullString
	var fspHighlightsStr, stackStr, reasonsStr string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&d.UserID, &d.DisplayName, &d.CategoryID, &d.SpecializationID, &d.SpecializationName,
		&d.GradeID, &d.GradeName, &d.GradeRank, &testScore, &d.HasFSP, &sportsRank,
		&d.FSPRating, &d.FSPScore, &d.FSPAchievementsCount, &bestPlace, &d.FSPWeightSum,
		&fspHighlightsStr, &stackStr, &yearsExp, &location, &d.PeriodicTasksSolved,
		&d.ActivityScore, &d.CalculatedScore, &d.Explanation, &reasonsStr, &d.LastActiveAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get candidate doc: %w", err)
	}

	if testScore.Valid {
		d.TestScore = &testScore.Float64
	}
	if bestPlace.Valid {
		bp := int(bestPlace.Int32)
		d.FSPBestPlace = &bp
	}
	if yearsExp.Valid {
		d.YearsExperience = &yearsExp.Float64
	}
	if location.Valid {
		d.Location = &location.String
	}
	if sportsRank.Valid {
		d.SportsRank = sportsRank.String
	}

	d.FSPHighlights = parseStringArray(fspHighlightsStr)
	d.Stack = parseUUIDArray(stackStr)
	d.Reasons = parseStringArray(reasonsStr)

	return &d, nil
}

func (r *PostgresRepository) GetCategoryRanking(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.CandidateSearchDoc, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT user_id, display_name, category_id, specialization_id, specialization_name,
		       grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
		       fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
		       fsp_highlights::text, stack::text, years_experience, location, periodic_tasks_solved,
		       activity_score, calculated_score, explanation, reasons::text, last_active_at, updated_at
		FROM candidate_search_docs
		WHERE category_id = $1
		ORDER BY calculated_score DESC, test_score DESC NULLS LAST, fsp_score DESC
		LIMIT $2;
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("query category ranking: %w", err)
	}
	defer rows.Close()

	var docs []domain.CandidateSearchDoc
	for rows.Next() {
		var d domain.CandidateSearchDoc
		var testScore sql.NullFloat64
		var bestPlace sql.NullInt32
		var yearsExp sql.NullFloat64
		var location sql.NullString
		var sportsRank sql.NullString
		var fspHighlightsStr, stackStr, reasonsStr string

		err := rows.Scan(
			&d.UserID, &d.DisplayName, &d.CategoryID, &d.SpecializationID, &d.SpecializationName,
			&d.GradeID, &d.GradeName, &d.GradeRank, &testScore, &d.HasFSP, &sportsRank,
			&d.FSPRating, &d.FSPScore, &d.FSPAchievementsCount, &bestPlace, &d.FSPWeightSum,
			&fspHighlightsStr, &stackStr, &yearsExp, &location, &d.PeriodicTasksSolved,
			&d.ActivityScore, &d.CalculatedScore, &d.Explanation, &reasonsStr, &d.LastActiveAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ranking doc: %w", err)
		}

		if testScore.Valid {
			d.TestScore = &testScore.Float64
		}
		if bestPlace.Valid {
			bp := int(bestPlace.Int32)
			d.FSPBestPlace = &bp
		}
		if yearsExp.Valid {
			d.YearsExperience = &yearsExp.Float64
		}
		if location.Valid {
			d.Location = &location.String
		}
		if sportsRank.Valid {
			d.SportsRank = sportsRank.String
		}

		d.FSPHighlights = parseStringArray(fspHighlightsStr)
		d.Stack = parseUUIDArray(stackStr)
		d.Reasons = parseStringArray(reasonsStr)

		docs = append(docs, d)
	}

	return docs, nil
}

func (r *PostgresRepository) CalculateTestPercentile(ctx context.Context, categoryID uuid.UUID, testScore float64) (float64, error) {
	query := `
		SELECT 
		  COUNT(CASE WHEN test_score <= $2 THEN 1 END)::float / NULLIF(COUNT(*), 0)::float * 100.0
		FROM candidate_search_docs
		WHERE category_id = $1 AND test_score IS NOT NULL;
	`
	var percentile sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, categoryID, testScore).Scan(&percentile)
	if err != nil || !percentile.Valid {
		// Fallback heuristics
		if testScore >= 95 {
			return 98.0, nil
		} else if testScore >= 90 {
			return 92.0, nil
		} else if testScore >= 80 {
			return 80.0, nil
		}
		return 60.0, nil
	}
	return percentile.Float64, nil
}

func (r *PostgresRepository) Upsert(ctx context.Context, doc *domain.CandidateSearchDoc) error {
	query := `
		INSERT INTO candidate_search_docs (
			user_id, display_name, category_id, specialization_id, specialization_name,
			grade_id, grade_name, grade_rank, test_score, has_fsp, sports_rank,
			fsp_rating, fsp_score, fsp_achievements_count, fsp_best_place, fsp_weight_sum,
			fsp_highlights, stack, years_experience, location, periodic_tasks_solved,
			activity_score, calculated_score, explanation, reasons, last_active_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17::text[], $18::uuid[], $19, $20, $21,
			$22, $23, $24, $25::text[], $26, $27
		)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name           = EXCLUDED.display_name,
			category_id            = EXCLUDED.category_id,
			specialization_id      = EXCLUDED.specialization_id,
			specialization_name    = EXCLUDED.specialization_name,
			grade_id               = EXCLUDED.grade_id,
			grade_name             = EXCLUDED.grade_name,
			grade_rank             = EXCLUDED.grade_rank,
			test_score             = EXCLUDED.test_score,
			has_fsp                = EXCLUDED.has_fsp,
			sports_rank            = EXCLUDED.sports_rank,
			fsp_rating             = EXCLUDED.fsp_rating,
			fsp_score              = EXCLUDED.fsp_score,
			fsp_achievements_count = EXCLUDED.fsp_achievements_count,
			fsp_best_place         = EXCLUDED.fsp_best_place,
			fsp_weight_sum         = EXCLUDED.fsp_weight_sum,
			fsp_highlights         = EXCLUDED.fsp_highlights,
			stack                  = EXCLUDED.stack,
			years_experience       = EXCLUDED.years_experience,
			location               = EXCLUDED.location,
			periodic_tasks_solved  = EXCLUDED.periodic_tasks_solved,
			activity_score         = EXCLUDED.activity_score,
			calculated_score       = EXCLUDED.calculated_score,
			explanation            = EXCLUDED.explanation,
			reasons                = EXCLUDED.reasons,
			last_active_at         = EXCLUDED.last_active_at,
			updated_at             = EXCLUDED.updated_at;
	`

	now := time.Now()
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if doc.LastActiveAt.IsZero() {
		doc.LastActiveAt = now
	}

	highlightsArr := formatStringArray(doc.FSPHighlights)
	stackArr := formatUUIDArray(doc.Stack)
	reasonsArr := formatStringArray(doc.Reasons)

	_, err := r.db.ExecContext(ctx, query,
		doc.UserID, doc.DisplayName, doc.CategoryID, doc.SpecializationID, doc.SpecializationName,
		doc.GradeID, doc.GradeName, doc.GradeRank, doc.TestScore, doc.HasFSP, doc.SportsRank,
		doc.FSPRating, doc.FSPScore, doc.FSPAchievementsCount, doc.FSPBestPlace, doc.FSPWeightSum,
		highlightsArr, stackArr, doc.YearsExperience, doc.Location, doc.PeriodicTasksSolved,
		doc.ActivityScore, doc.CalculatedScore, doc.Explanation, reasonsArr, doc.LastActiveAt, doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert candidate search doc: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM candidate_search_docs").Scan(&count)
	return count, err
}

// ----------------------------------------------------
// Postgres Array Parsing Helpers
// ----------------------------------------------------

func parseStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "NULL" {
		return []string{}
	}
	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")
	if raw == "" {
		return []string{}
	}

	// Simple CSV parse with quote stripping
	var res []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			res = append(res, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		res = append(res, current.String())
	}
	return res
}

func parseUUIDArray(raw string) []uuid.UUID {
	strArr := parseStringArray(raw)
	res := make([]uuid.UUID, 0, len(strArr))
	for _, s := range strArr {
		s = strings.TrimSpace(s)
		if u, err := uuid.Parse(s); err == nil {
			res = append(res, u)
		}
	}
	return res
}

func formatStringArray(items []string) string {
	if len(items) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	for i, item := range items {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Quote(item))
	}
	sb.WriteString("}")
	return sb.String()
}

func formatUUIDArray(items []uuid.UUID) string {
	if len(items) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	for i, item := range items {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(item.String())
	}
	sb.WriteString("}")
	return sb.String()
}
