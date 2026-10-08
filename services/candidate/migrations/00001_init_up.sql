CREATE TABLE candidate_profiles (
  user_id             UUID PRIMARY KEY, -- = auth.users.id
  first_name          TEXT,
  last_name           TEXT,
  middle_name         TEXT,
  headline            TEXT,
  about               TEXT,
  location            TEXT,
  years_experience    NUMERIC(3,1),
  current_category_id UUID,
  current_grade_id    UUID,
  specialization_id   UUID,
  fsp_member_id       TEXT, -- nullable, ок если нет
  fsp_linked_at       TIMESTAMPTZ,
  visibility          JSONB NOT NULL DEFAULT '{}'::jsonb, -- {contacts:true, fsp:true, ...}
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_candidate_profiles_category ON candidate_profiles(current_category_id);
CREATE INDEX idx_candidate_profiles_specialization ON candidate_profiles(specialization_id);
CREATE INDEX idx_candidate_profiles_grade ON candidate_profiles(current_grade_id);

-- PII вынесены отдельно
CREATE TABLE candidate_contacts (
  user_id    UUID PRIMARY KEY REFERENCES candidate_profiles(user_id) ON DELETE CASCADE,
  email      CITEXT,
  phone      TEXT,
  telegram   TEXT,
  github     TEXT,
  linkedin   TEXT,
  website    TEXT,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE resumes (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL,
  title      TEXT NOT NULL,
  source     TEXT NOT NULL CHECK (source IN ('manual','pdf_upload','generated')),
  file_path  TEXT,
  parsed     JSONB,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_resumes_user ON resumes(user_id);

CREATE TABLE candidate_technologies (
  user_id       UUID NOT NULL,
  technology_id UUID NOT NULL,
  level         SMALLINT CHECK (level BETWEEN 1 AND 5),
  PRIMARY KEY (user_id, technology_id)
);
CREATE INDEX idx_candidate_tech_tech ON candidate_technologies(technology_id);

CREATE TABLE candidate_experiences (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  company     TEXT NOT NULL,
  position    TEXT NOT NULL,
  started_at  DATE NOT NULL,
  ended_at    DATE,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_candidate_exp_user ON candidate_experiences(user_id);

-- ключевая таблица для ранжирования по ФСП
CREATE TABLE fsp_achievements (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  external_id TEXT, -- из реестра ФСП, nullable
  event_name  TEXT NOT NULL,
  event_date  DATE,
  place       SMALLINT,
  category    TEXT,
  score       NUMERIC(10,2),
  weight      SMALLINT NOT NULL DEFAULT 1, -- важность в ранжировании
  payload     JSONB NOT NULL DEFAULT '{}'::jsonb,
  source      TEXT NOT NULL DEFAULT 'mock'
                CHECK (source IN ('mock','api','manual')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, external_id)
);
CREATE INDEX idx_fsp_user ON fsp_achievements(user_id);
CREATE INDEX idx_fsp_user_weight ON fsp_achievements(user_id, weight DESC);

-- история смены категории / грейда
CREATE TABLE candidate_category_history (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID NOT NULL,
  category_id       UUID NOT NULL,
  grade_id          UUID NOT NULL,
  specialization_id UUID NOT NULL,
  test_session_id   UUID,
  reason            TEXT NOT NULL CHECK (reason IN ('initial_test','retake','manual')),
  effective_from    TIMESTAMPTZ NOT NULL DEFAULT now(),
  effective_to      TIMESTAMPTZ
);
CREATE INDEX idx_cat_hist_user ON candidate_category_history(user_id, effective_from DESC);