-- +goose Up

-- Таблица грейдов с рангами для автономного расчета переходов (upgrade/downgrade)
CREATE TABLE IF NOT EXISTS grade_definitions (
  id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code  TEXT UNIQUE NOT NULL,
  name  TEXT NOT NULL,
  rank  SMALLINT NOT NULL UNIQUE -- 1..7
);

-- Таблица категорий со связкой со специализацией и грейдом
CREATE TABLE IF NOT EXISTS category_definitions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  specialization_id UUID NOT NULL,
  grade_id          UUID NOT NULL REFERENCES grade_definitions(id),
  slug              TEXT UNIQUE NOT NULL,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (specialization_id, grade_id)
);

-- Заполнение базовых грейдов с фиксированными детерминированными UUID
INSERT INTO grade_definitions (id, code, name, rank) VALUES
  ('00000000-0000-0000-0001-000000000001', 'intern',      'Intern',   1),
  ('00000000-0000-0000-0001-000000000002', 'junior',      'Junior',   2),
  ('00000000-0000-0000-0001-000000000003', 'junior_plus', 'Junior+',  3),
  ('00000000-0000-0000-0001-000000000004', 'middle',      'Middle',   4),
  ('00000000-0000-0000-0001-000000000005', 'middle_plus', 'Middle+',  5),
  ('00000000-0000-0000-0001-000000000006', 'senior',      'Senior',   6),
  ('00000000-0000-0000-0001-000000000007', 'lead',        'Lead',     7)
ON CONFLICT (code) DO UPDATE SET rank = EXCLUDED.rank, name = EXCLUDED.name;

-- Базовая специализация Backend
INSERT INTO category_definitions (id, specialization_id, grade_id, slug) VALUES
  ('00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000001', 'backend_intern'),
  ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000002', 'backend_junior'),
  ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000003', 'backend_junior_plus'),
  ('00000000-0000-0000-0002-000000000004', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000004', 'backend_middle'),
  ('00000000-0000-0000-0002-000000000005', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000005', 'backend_middle_plus'),
  ('00000000-0000-0000-0002-000000000006', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000006', 'backend_senior'),
  ('00000000-0000-0000-0002-000000000007', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0001-000000000007', 'backend_lead')
ON CONFLICT (slug) DO NOTHING;

-- Опросник на входе
CREATE TABLE IF NOT EXISTS candidate_questionnaires (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id            UUID NOT NULL,
  specialization_id  UUID NOT NULL,
  claimed_grade_id   UUID NOT NULL,
  target_category_id UUID NOT NULL,
  years_experience   NUMERIC(4,1) DEFAULT 0,
  technologies       JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_questionnaires_user ON candidate_questionnaires(user_id, created_at DESC);

-- Фиксация подтвержденной категории и грейда кандидата
CREATE TABLE IF NOT EXISTS candidate_category_state (
  user_id             UUID PRIMARY KEY,
  current_category_id UUID NOT NULL,
  current_grade_id    UUID NOT NULL REFERENCES grade_definitions(id),
  specialization_id   UUID,
  status              TEXT NOT NULL CHECK (status IN ('confirmed', 'downgrade_offered', 'upgrade_offered')),
  test_score          NUMERIC(6,3) NOT NULL,
  ability_estimate    NUMERIC(5,3),
  last_session_id     UUID NOT NULL REFERENCES test_sessions(id),
  can_change_at       TIMESTAMPTZ,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cat_state_grade ON candidate_category_state(current_grade_id);

-- +goose Down
DROP TABLE IF EXISTS candidate_category_state;
DROP TABLE IF EXISTS candidate_questionnaires;
DROP TABLE IF EXISTS category_definitions;
DROP TABLE IF EXISTS grade_definitions;
