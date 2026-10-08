CREATE TABLE test_templates (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id UUID NOT NULL,
  version     INT NOT NULL DEFAULT 1,
  config      JSONB NOT NULL, -- распределение сложностей, веса
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (category_id, version)
);

CREATE TABLE tasks (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id    UUID NOT NULL REFERENCES test_templates(id) ON DELETE CASCADE,
  type           TEXT NOT NULL
                   CHECK (type IN ('single_choice','multi_choice','code','text','sql','regex')),
  topic          TEXT NOT NULL,
  title          TEXT NOT NULL,
  body           TEXT NOT NULL,
  generator      JSONB, -- правила параметризации
  solution       JSONB, -- эталон/рубрика
  rubric         JSONB,
  difficulty     SMALLINT NOT NULL CHECK (difficulty BETWEEN 1 AND 5),
  discrimination NUMERIC(5,3), -- IRT a
  difficulty_irt NUMERIC(5,3), -- IRT b
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tasks_template ON tasks(template_id);

CREATE TABLE test_sessions (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               UUID NOT NULL,
  target_category_id    UUID NOT NULL, -- куда метит кандидат
  template_id           UUID NOT NULL REFERENCES test_templates(id),
  status                TEXT NOT NULL DEFAULT 'in_progress'
                          CHECK (status IN ('in_progress','submitted','evaluated','expired','cancelled')),
  ability_estimate      NUMERIC(5,3), -- theta IRT
  score                 NUMERIC(6,3),
  resulting_grade_id    UUID,
  resulting_category_id UUID,
  started_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at           TIMESTAMPTZ
);
CREATE INDEX idx_sessions_user ON test_sessions(user_id, started_at DESC);
CREATE INDEX idx_sessions_status ON test_sessions(status);

CREATE TABLE test_items (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id     UUID NOT NULL REFERENCES test_sessions(id) ON DELETE CASCADE,
  task_id        UUID NOT NULL REFERENCES tasks(id),
  position       INT NOT NULL,
  variant_params JSONB NOT NULL DEFAULT '{}'::jsonb,
  rendered_body  TEXT NOT NULL,
  status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending','answered','skipped')),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (session_id, position)
);

CREATE TABLE test_answers (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id     UUID NOT NULL REFERENCES test_items(id) ON DELETE CASCADE,
  answer      JSONB NOT NULL,
  is_correct  BOOLEAN,
  score       NUMERIC(5,3),
  graded_by   TEXT NOT NULL DEFAULT 'auto'
                CHECK (graded_by IN ('auto','ml','reviewer')),
  graded_at   TIMESTAMPTZ,
  answered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_answers_item ON test_answers(item_id);

-- лимит смены грейда
CREATE TABLE grade_change_events (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL,
  from_grade_id UUID,
  to_grade_id   UUID NOT NULL,
  reason        TEXT NOT NULL,
  changed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_grade_change_user ON grade_change_events(user_id, changed_at DESC);

-- периодические короткие задачи от работодателей
CREATE TABLE periodic_tasks (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employer_user_id UUID NOT NULL,
  category_id      UUID NOT NULL,
  title            TEXT NOT NULL,
  body             TEXT NOT NULL,
  is_active        BOOLEAN NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_periodic_cat ON periodic_tasks(category_id, is_active);

CREATE TABLE periodic_submissions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id    UUID NOT NULL REFERENCES periodic_tasks(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL,
  answer     TEXT NOT NULL,
  score      NUMERIC(5,3),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (task_id, user_id)
);