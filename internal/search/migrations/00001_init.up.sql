-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE candidate_search_docs (
  user_id                UUID PRIMARY KEY,
  category_id            UUID NOT NULL,
  specialization_id      UUID NOT NULL,
  grade_id               UUID NOT NULL,
  grade_rank             SMALLINT NOT NULL,
  test_score             NUMERIC(6,3),
  fsp_achievements_count INT NOT NULL DEFAULT 0,
  fsp_best_place         SMALLINT,
  fsp_weight_sum         INT NOT NULL DEFAULT 0,
  stack                  UUID[] NOT NULL DEFAULT '{}',
  years_experience       NUMERIC(3,1),
  location               TEXT,
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- основной индекс для ранжирования внутри категории
CREATE INDEX idx_search_rank ON candidate_search_docs
  (category_id, fsp_weight_sum DESC, test_score DESC NULLS LAST);

CREATE INDEX idx_search_stack ON candidate_search_docs USING GIN (stack);
CREATE INDEX idx_search_spec  ON candidate_search_docs(specialization_id, grade_rank);