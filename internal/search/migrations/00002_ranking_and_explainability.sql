-- +goose Up
ALTER TABLE candidate_search_docs
  ADD COLUMN IF NOT EXISTS display_name           TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS specialization_name    TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS grade_name             TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS has_fsp                BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS sports_rank            TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS fsp_rating             INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fsp_score              NUMERIC(6,3) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS fsp_highlights         TEXT[] NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS periodic_tasks_solved  INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS activity_score         NUMERIC(6,3) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS calculated_score       NUMERIC(6,3) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS explanation            TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS reasons                TEXT[] NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS last_active_at         TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_search_category_score 
  ON candidate_search_docs(category_id, calculated_score DESC);

CREATE INDEX IF NOT EXISTS idx_search_calculated_score 
  ON candidate_search_docs(calculated_score DESC);

CREATE INDEX IF NOT EXISTS idx_search_has_fsp 
  ON candidate_search_docs(has_fsp);

-- +goose Down
DROP INDEX IF EXISTS idx_search_has_fsp;
DROP INDEX IF EXISTS idx_search_calculated_score;
DROP INDEX IF EXISTS idx_search_category_score;

ALTER TABLE candidate_search_docs
  DROP COLUMN IF EXISTS last_active_at,
  DROP COLUMN IF EXISTS reasons,
  DROP COLUMN IF EXISTS explanation,
  DROP COLUMN IF EXISTS calculated_score,
  DROP COLUMN IF EXISTS activity_score,
  DROP COLUMN IF EXISTS periodic_tasks_solved,
  DROP COLUMN IF EXISTS fsp_highlights,
  DROP COLUMN IF EXISTS fsp_score,
  DROP COLUMN IF EXISTS fsp_rating,
  DROP COLUMN IF EXISTS sports_rank,
  DROP COLUMN IF EXISTS has_fsp,
  DROP COLUMN IF EXISTS grade_name,
  DROP COLUMN IF EXISTS specialization_name,
  DROP COLUMN IF EXISTS display_name;
