package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"TimofeyChernyshev/LCT-hack-2/internal/testing/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyAnswer = errors.New("item already answered")
)

type Repository interface {
	GetTemplateForCategory(ctx context.Context, categoryID uuid.UUID) (*domain.Template, error)
	GetTasksForTemplate(ctx context.Context, templateID uuid.UUID, limit int) ([]domain.Task, error)
	CreateSession(ctx context.Context, session *domain.Session, items []domain.SessionItem) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	GetSessionItemWithTask(ctx context.Context, sessionID, itemID uuid.UUID) (*domain.SessionItem, *domain.Task, error)
	SaveAnswer(ctx context.Context, answer *domain.TestAnswer, status domain.SessionItemStatus) error
	UpdateSessionResult(ctx context.Context, session *domain.Session) error
	GetSessionEvaluations(ctx context.Context, sessionID uuid.UUID) ([]SessionItemEval, error)

	GetLatestGradeChange(ctx context.Context, userID uuid.UUID) (*domain.GradeChangeEvent, error)
	ListGradeChanges(ctx context.Context, userID uuid.UUID) ([]domain.GradeChangeEvent, error)
	CreateGradeChangeEvent(ctx context.Context, event *domain.GradeChangeEvent) error

	ListPeriodicTasks(ctx context.Context, categoryID *uuid.UUID) ([]domain.PeriodicTask, error)
	CreatePeriodicTask(ctx context.Context, task *domain.PeriodicTask) error
	CreatePeriodicSubmission(ctx context.Context, sub *domain.PeriodicSubmission) error
	GetCandidateCategory(ctx context.Context, userID uuid.UUID) (categoryID, gradeID *uuid.UUID, score *float64, err error)

	// BE2-02: Grade scale, questionnaire, and category state
	GetGradeDefinition(ctx context.Context, gradeID uuid.UUID) (*domain.GradeDefinition, error)
	GetGradeByRank(ctx context.Context, rank int) (*domain.GradeDefinition, error)
	FindCategory(ctx context.Context, specializationID, gradeID uuid.UUID) (*domain.CategoryDefinition, error)
	GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*domain.CategoryDefinition, error)
	SaveQuestionnaire(ctx context.Context, q *domain.CandidateQuestionnaire) error
	GetLatestQuestionnaire(ctx context.Context, userID uuid.UUID) (*domain.CandidateQuestionnaire, error)
	SaveCandidateCategoryState(ctx context.Context, state *domain.CandidateCategoryState) error
	GetCandidateCategoryState(ctx context.Context, userID uuid.UUID) (*domain.CandidateCategoryState, error)

	// BE2-03: FSP Registry Integration & Candidate FSP State
	SaveCandidateFSP(ctx context.Context, profile *domain.CandidateFSPProfile) error
	GetCandidateFSP(ctx context.Context, userID uuid.UUID) (*domain.CandidateFSPProfile, error)
	UnlinkCandidateFSP(ctx context.Context, userID uuid.UUID) error
}

type SessionItemEval struct {
	Item           domain.SessionItem
	Task           domain.Task
	Answer         *domain.TestAnswer
	Discrimination float64
	Difficulty     int
	DifficultyIRT  float64
	Score          float64
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetTemplateForCategory(ctx context.Context, categoryID uuid.UUID) (*domain.Template, error) {
	query := `
		SELECT id, category_id, version, config, is_active, created_at
		FROM test_templates
		WHERE category_id = $1 AND is_active = TRUE
		ORDER BY version DESC
		LIMIT 1;
	`
	var t domain.Template
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&t.ID, &t.CategoryID, &t.Version, &t.Config, &t.IsActive, &t.CreatedAt,
	)
	if err == nil {
		return &t, nil
	}

	// Fallback: pick any active template if exact category template not found
	fallbackQuery := `
		SELECT id, category_id, version, config, is_active, created_at
		FROM test_templates
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1;
	`
	err = r.db.QueryRowContext(ctx, fallbackQuery).Scan(
		&t.ID, &t.CategoryID, &t.Version, &t.Config, &t.IsActive, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query template: %w", err)
	}
	return &t, nil
}

func (r *PostgresRepository) GetTasksForTemplate(ctx context.Context, templateID uuid.UUID, limit int) ([]domain.Task, error) {
	query := `
		SELECT id, template_id, type, topic, title, body, generator, solution, rubric,
		       difficulty, COALESCE(discrimination, 1.0), COALESCE(difficulty_irt, 0.0), is_active, created_at
		FROM tasks
		WHERE (template_id = $1 OR is_active = TRUE)
		ORDER BY RANDOM()
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, templateID, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		var generator, solution, rubric []byte
		err := rows.Scan(
			&t.ID, &t.TemplateID, &t.Type, &t.Topic, &t.Title, &t.Body,
			&generator, &solution, &rubric,
			&t.Difficulty, &t.Discrimination, &t.DifficultyIRT, &t.IsActive, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		t.Generator = generator
		t.Solution = solution
		t.Rubric = rubric
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session *domain.Session, items []domain.SessionItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	sessQuery := `
		INSERT INTO test_sessions (id, user_id, target_category_id, template_id, status, started_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err = tx.ExecContext(ctx, sessQuery,
		session.ID, session.UserID, session.TargetCategoryID, session.TemplateID,
		session.Status, session.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	itemQuery := `
		INSERT INTO test_items (id, session_id, task_id, position, variant_params, rendered_body, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	for _, it := range items {
		_, err = tx.ExecContext(ctx, itemQuery,
			it.ID, session.ID, it.TaskID, it.Position, it.VariantParams, it.RenderedBody, it.Status, it.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert item %d: %w", it.Position, err)
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, target_category_id, template_id, status, ability_estimate, score,
		       resulting_grade_id, resulting_category_id, started_at, finished_at
		FROM test_sessions
		WHERE id = $1;
	`
	var s domain.Session
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.TargetCategoryID, &s.TemplateID, &s.Status,
		&s.AbilityEstimate, &s.Score, &s.ResultingGradeID, &s.ResultingCategoryID,
		&s.StartedAt, &s.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query session: %w", err)
	}

	// Fetch items
	itemsQuery := `
		SELECT id, session_id, task_id, position, variant_params, rendered_body, status, created_at
		FROM test_items
		WHERE session_id = $1
		ORDER BY position ASC;
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query session items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it domain.SessionItem
		err := rows.Scan(
			&it.ID, &it.SessionID, &it.TaskID, &it.Position, &it.VariantParams,
			&it.RenderedBody, &it.Status, &it.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		s.Items = append(s.Items, it)
	}

	return &s, rows.Err()
}

func (r *PostgresRepository) ListSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	query := `
		SELECT id, user_id, target_category_id, template_id, status, ability_estimate, score,
		       resulting_grade_id, resulting_category_id, started_at, finished_at
		FROM test_sessions
		WHERE user_id = $1
		ORDER BY started_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		err := rows.Scan(
			&s.ID, &s.UserID, &s.TargetCategoryID, &s.TemplateID, &s.Status,
			&s.AbilityEstimate, &s.Score, &s.ResultingGradeID, &s.ResultingCategoryID,
			&s.StartedAt, &s.FinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (r *PostgresRepository) GetSessionItemWithTask(ctx context.Context, sessionID, itemID uuid.UUID) (*domain.SessionItem, *domain.Task, error) {
	query := `
		SELECT i.id, i.session_id, i.task_id, i.position, i.variant_params, i.rendered_body, i.status, i.created_at,
		       t.id, t.template_id, t.type, t.topic, t.title, t.body, t.generator, t.solution, t.rubric,
		       t.difficulty, COALESCE(t.discrimination, 1.0), COALESCE(t.difficulty_irt, 0.0), t.is_active, t.created_at
		FROM test_items i
		JOIN tasks t ON i.task_id = t.id
		WHERE i.session_id = $1 AND i.id = $2;
	`
	var it domain.SessionItem
	var t domain.Task
	var generator, solution, rubric []byte

	err := r.db.QueryRowContext(ctx, query, sessionID, itemID).Scan(
		&it.ID, &it.SessionID, &it.TaskID, &it.Position, &it.VariantParams, &it.RenderedBody, &it.Status, &it.CreatedAt,
		&t.ID, &t.TemplateID, &t.Type, &t.Topic, &t.Title, &t.Body, &generator, &solution, &rubric,
		&t.Difficulty, &t.Discrimination, &t.DifficultyIRT, &t.IsActive, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("query item with task: %w", err)
	}

	t.Generator = generator
	t.Solution = solution
	t.Rubric = rubric
	it.Task = &t

	return &it, &t, nil
}

func (r *PostgresRepository) SaveAnswer(ctx context.Context, answer *domain.TestAnswer, status domain.SessionItemStatus) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Check if already answered
	var count int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_answers WHERE item_id = $1", answer.ItemID).Scan(&count)
	if err != nil {
		return fmt.Errorf("check existing answer: %w", err)
	}
	if count > 0 {
		return ErrAlreadyAnswer
	}

	ansQuery := `
		INSERT INTO test_answers (id, item_id, answer, is_correct, score, graded_by, graded_at, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err = tx.ExecContext(ctx, ansQuery,
		answer.ID, answer.ItemID, answer.Answer, answer.IsCorrect, answer.Score,
		answer.GradedBy, answer.GradedAt, answer.AnsweredAt,
	)
	if err != nil {
		return fmt.Errorf("insert answer: %w", err)
	}

	itemUpdate := `UPDATE test_items SET status = $1 WHERE id = $2;`
	_, err = tx.ExecContext(ctx, itemUpdate, status, answer.ItemID)
	if err != nil {
		return fmt.Errorf("update item status: %w", err)
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateSessionResult(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE test_sessions
		SET status = $1, ability_estimate = $2, score = $3, resulting_grade_id = $4,
		    resulting_category_id = $5, finished_at = $6
		WHERE id = $7;
	`
	_, err := r.db.ExecContext(ctx, query,
		session.Status, session.AbilityEstimate, session.Score,
		session.ResultingGradeID, session.ResultingCategoryID, session.FinishedAt,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("update session result: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetSessionEvaluations(ctx context.Context, sessionID uuid.UUID) ([]SessionItemEval, error) {
	query := `
		SELECT i.id, i.session_id, i.task_id, i.position, i.status,
		       t.id, t.type, t.difficulty, COALESCE(t.discrimination, 1.0), COALESCE(t.difficulty_irt, 0.0),
		       COALESCE(a.score, 0.0), a.is_correct
		FROM test_items i
		JOIN tasks t ON i.task_id = t.id
		LEFT JOIN test_answers a ON i.id = a.item_id
		WHERE i.session_id = $1
		ORDER BY i.position ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query session evaluations: %w", err)
	}
	defer rows.Close()

	var evals []SessionItemEval
	for rows.Next() {
		var eval SessionItemEval
		var isCorrect sql.NullBool
		err := rows.Scan(
			&eval.Item.ID, &eval.Item.SessionID, &eval.Item.TaskID, &eval.Item.Position, &eval.Item.Status,
			&eval.Task.ID, &eval.Task.Type, &eval.Difficulty, &eval.Discrimination, &eval.DifficultyIRT,
			&eval.Score, &isCorrect,
		)
		if err != nil {
			return nil, fmt.Errorf("scan eval: %w", err)
		}
		evals = append(evals, eval)
	}
	return evals, rows.Err()
}

func (r *PostgresRepository) GetLatestGradeChange(ctx context.Context, userID uuid.UUID) (*domain.GradeChangeEvent, error) {
	query := `
		SELECT id, user_id, from_grade_id, to_grade_id, reason, changed_at
		FROM grade_change_events
		WHERE user_id = $1
		ORDER BY changed_at DESC
		LIMIT 1;
	`
	var e domain.GradeChangeEvent
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&e.ID, &e.UserID, &e.FromGradeID, &e.ToGradeID, &e.Reason, &e.ChangedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest grade change: %w", err)
	}
	return &e, nil
}

func (r *PostgresRepository) ListGradeChanges(ctx context.Context, userID uuid.UUID) ([]domain.GradeChangeEvent, error) {
	query := `
		SELECT id, user_id, from_grade_id, to_grade_id, reason, changed_at
		FROM grade_change_events
		WHERE user_id = $1
		ORDER BY changed_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query grade changes: %w", err)
	}
	defer rows.Close()

	var events []domain.GradeChangeEvent
	for rows.Next() {
		var e domain.GradeChangeEvent
		err := rows.Scan(
			&e.ID, &e.UserID, &e.FromGradeID, &e.ToGradeID, &e.Reason, &e.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan grade change: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) CreateGradeChangeEvent(ctx context.Context, event *domain.GradeChangeEvent) error {
	query := `
		INSERT INTO grade_change_events (id, user_id, from_grade_id, to_grade_id, reason, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.FromGradeID, event.ToGradeID, event.Reason, event.ChangedAt,
	)
	if err != nil {
		return fmt.Errorf("insert grade change: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListPeriodicTasks(ctx context.Context, categoryID *uuid.UUID) ([]domain.PeriodicTask, error) {
	query := `
		SELECT id, employer_user_id, category_id, title, body, is_active, created_at
		FROM periodic_tasks
		WHERE is_active = TRUE AND ($1::uuid IS NULL OR category_id = $1)
		ORDER BY created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("query periodic tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.PeriodicTask
	for rows.Next() {
		var t domain.PeriodicTask
		err := rows.Scan(
			&t.ID, &t.EmployerUserID, &t.CategoryID, &t.Title, &t.Body, &t.IsActive, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan periodic task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *PostgresRepository) CreatePeriodicTask(ctx context.Context, task *domain.PeriodicTask) error {
	query := `
		INSERT INTO periodic_tasks (id, employer_user_id, category_id, title, body, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.EmployerUserID, task.CategoryID, task.Title, task.Body, task.IsActive, task.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert periodic task: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreatePeriodicSubmission(ctx context.Context, sub *domain.PeriodicSubmission) error {
	query := `
		INSERT INTO periodic_submissions (id, task_id, user_id, answer, score, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (task_id, user_id) DO UPDATE SET answer = EXCLUDED.answer, score = EXCLUDED.score;
	`
	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.TaskID, sub.UserID, sub.Answer, sub.Score, sub.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert periodic submission: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetCandidateCategory(ctx context.Context, userID uuid.UUID) (categoryID, gradeID *uuid.UUID, score *float64, err error) {
	st, err := r.GetCandidateCategoryState(ctx, userID)
	if err == nil && st != nil {
		return &st.CurrentCategoryID, &st.CurrentGradeID, &st.TestScore, nil
	}

	query := `
		SELECT target_category_id, resulting_grade_id, score
		FROM test_sessions
		WHERE user_id = $1 AND status IN ('evaluated', 'submitted')
		ORDER BY finished_at DESC NULLS LAST, started_at DESC
		LIMIT 1;
	`
	var catID, grdID uuid.NullUUID
	var sc sql.NullFloat64
	err = r.db.QueryRowContext(ctx, query, userID).Scan(&catID, &grdID, &sc)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, fmt.Errorf("query candidate category: %w", err)
	}

	if catID.Valid {
		categoryID = &catID.UUID
	}
	if grdID.Valid {
		gradeID = &grdID.UUID
	}
	if sc.Valid {
		score = &sc.Float64
	}
	return categoryID, gradeID, score, nil
}

func (r *PostgresRepository) GetGradeDefinition(ctx context.Context, gradeID uuid.UUID) (*domain.GradeDefinition, error) {
	query := `SELECT id, code, name, rank FROM grade_definitions WHERE id = $1;`
	var g domain.GradeDefinition
	err := r.db.QueryRowContext(ctx, query, gradeID).Scan(&g.ID, &g.Code, &g.Name, &g.Rank)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query grade: %w", err)
	}
	return &g, nil
}

func (r *PostgresRepository) GetGradeByRank(ctx context.Context, rank int) (*domain.GradeDefinition, error) {
	query := `SELECT id, code, name, rank FROM grade_definitions WHERE rank = $1;`
	var g domain.GradeDefinition
	err := r.db.QueryRowContext(ctx, query, rank).Scan(&g.ID, &g.Code, &g.Name, &g.Rank)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query grade by rank: %w", err)
	}
	return &g, nil
}

func (r *PostgresRepository) FindCategory(ctx context.Context, specializationID, gradeID uuid.UUID) (*domain.CategoryDefinition, error) {
	query := `SELECT id, specialization_id, grade_id, slug, is_active FROM category_definitions WHERE specialization_id = $1 AND grade_id = $2 LIMIT 1;`
	var c domain.CategoryDefinition
	err := r.db.QueryRowContext(ctx, query, specializationID, gradeID).Scan(&c.ID, &c.SpecializationID, &c.GradeID, &c.Slug, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find category: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepository) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*domain.CategoryDefinition, error) {
	query := `SELECT id, specialization_id, grade_id, slug, is_active FROM category_definitions WHERE id = $1;`
	var c domain.CategoryDefinition
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&c.ID, &c.SpecializationID, &c.GradeID, &c.Slug, &c.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query category by id: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepository) SaveQuestionnaire(ctx context.Context, q *domain.CandidateQuestionnaire) error {
	query := `
		INSERT INTO candidate_questionnaires (id, user_id, specialization_id, claimed_grade_id, target_category_id, years_experience, technologies, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	_, err := r.db.ExecContext(ctx, query,
		q.ID, q.UserID, q.SpecializationID, q.ClaimedGradeID, q.TargetCategoryID,
		q.YearsExperience, q.Technologies, q.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert questionnaire: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetLatestQuestionnaire(ctx context.Context, userID uuid.UUID) (*domain.CandidateQuestionnaire, error) {
	query := `
		SELECT id, user_id, specialization_id, claimed_grade_id, target_category_id, years_experience, technologies, created_at
		FROM candidate_questionnaires
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`
	var q domain.CandidateQuestionnaire
	var techBytes []byte
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&q.ID, &q.UserID, &q.SpecializationID, &q.ClaimedGradeID, &q.TargetCategoryID,
		&q.YearsExperience, &techBytes, &q.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest questionnaire: %w", err)
	}
	q.Technologies = techBytes
	return &q, nil
}

func (r *PostgresRepository) SaveCandidateCategoryState(ctx context.Context, state *domain.CandidateCategoryState) error {
	query := `
		INSERT INTO candidate_category_state (
			user_id, current_category_id, current_grade_id, specialization_id,
			status, test_score, ability_estimate, last_session_id, can_change_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id) DO UPDATE SET
			current_category_id = EXCLUDED.current_category_id,
			current_grade_id    = EXCLUDED.current_grade_id,
			specialization_id   = EXCLUDED.specialization_id,
			status              = EXCLUDED.status,
			test_score          = EXCLUDED.test_score,
			ability_estimate    = EXCLUDED.ability_estimate,
			last_session_id     = EXCLUDED.last_session_id,
			can_change_at       = EXCLUDED.can_change_at,
			updated_at          = EXCLUDED.updated_at;
	`
	var specID *uuid.UUID
	if state.SpecializationID != nil {
		specID = state.SpecializationID
	}
	_, err := r.db.ExecContext(ctx, query,
		state.UserID, state.CurrentCategoryID, state.CurrentGradeID, specID,
		state.Status, state.TestScore, state.AbilityEstimate, state.LastSessionID,
		state.CanChangeAt, state.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save candidate category state: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetCandidateCategoryState(ctx context.Context, userID uuid.UUID) (*domain.CandidateCategoryState, error) {
	query := `
		SELECT user_id, current_category_id, current_grade_id, specialization_id,
		       status, test_score, ability_estimate, last_session_id, can_change_at, updated_at
		FROM candidate_category_state
		WHERE user_id = $1;
	`
	var s domain.CandidateCategoryState
	var specID uuid.NullUUID
	var canChangeAt sql.NullTime
	var abilityEst sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.UserID, &s.CurrentCategoryID, &s.CurrentGradeID, &specID,
		&s.Status, &s.TestScore, &abilityEst, &s.LastSessionID, &canChangeAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query candidate category state: %w", err)
	}

	if specID.Valid {
		s.SpecializationID = &specID.UUID
	}
	if canChangeAt.Valid {
		s.CanChangeAt = &canChangeAt.Time
	}
	if abilityEst.Valid {
		s.AbilityEstimate = &abilityEst.Float64
	}
	return &s, nil
}
