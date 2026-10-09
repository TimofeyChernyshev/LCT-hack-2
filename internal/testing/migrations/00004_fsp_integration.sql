-- +goose Up

-- Профиль связки кандидата с реестром ФСП (обогащенные данные и скоринг)
CREATE TABLE IF NOT EXISTS candidate_fsp_profiles (
  user_id             UUID PRIMARY KEY,
  fsp_member_id       TEXT,
  full_name           TEXT,
  sports_rank         TEXT,
  fsp_rating          INT NOT NULL DEFAULT 0,
  region              TEXT,
  discipline          TEXT,
  has_fsp             BOOLEAN NOT NULL DEFAULT FALSE,
  fsp_score           NUMERIC(5,2) NOT NULL DEFAULT 0.00,
  fsp_weight_sum      INT NOT NULL DEFAULT 0,
  achievements_count  INT NOT NULL DEFAULT 0,
  best_place          INT,
  verification_source TEXT NOT NULL DEFAULT 'mock',
  linked_at           TIMESTAMPTZ,
  explanation         TEXT NOT NULL DEFAULT 'История участия в соревнованиях ФСП не привязана. Кандидат оценивается по результатам тестов и стеку компетенций.',
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_candidate_fsp_has_fsp ON candidate_fsp_profiles(has_fsp);
CREATE INDEX IF NOT EXISTS idx_candidate_fsp_score ON candidate_fsp_profiles(fsp_score DESC);
CREATE INDEX IF NOT EXISTS idx_candidate_fsp_member_id ON candidate_fsp_profiles(fsp_member_id);

-- Спортивные достижения кандидата из реестра ФСП
CREATE TABLE IF NOT EXISTS candidate_fsp_achievements (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES candidate_fsp_profiles(user_id) ON DELETE CASCADE,
  external_id TEXT,
  event_name  TEXT NOT NULL,
  event_date  DATE,
  place       SMALLINT,
  category    TEXT,
  score       NUMERIC(10,2),
  weight      SMALLINT NOT NULL DEFAULT 1,
  badge       TEXT,
  description TEXT,
  payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_fsp_achievements_user ON candidate_fsp_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_fsp_achievements_weight ON candidate_fsp_achievements(user_id, weight DESC);

-- +goose Down
DROP TABLE IF EXISTS candidate_fsp_achievements;
DROP TABLE IF EXISTS candidate_fsp_profiles;
