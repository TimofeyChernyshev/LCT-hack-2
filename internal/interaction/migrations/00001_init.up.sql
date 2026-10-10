-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE invitations (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employer_user_id UUID NOT NULL,
  company_id       UUID NOT NULL,
  candidate_user_id UUID NOT NULL,
  vacancy_id       UUID,
  need_id          UUID,
  message          TEXT NOT NULL,
  salary_min       INT NOT NULL,
  salary_max       INT NOT NULL,
  currency         TEXT NOT NULL DEFAULT 'RUB',
  contact_channel  TEXT NOT NULL,
  employer_contact TEXT,
  status           TEXT NOT NULL DEFAULT 'sent'
                     CHECK (status IN ('sent','viewed','accepted','rejected','withdrawn','expired')),
  status_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (salary_min <= salary_max)
);
CREATE INDEX idx_inv_candidate ON invitations(candidate_user_id, status, created_at DESC);
CREATE INDEX idx_inv_employer  ON invitations(employer_user_id, status, created_at DESC);

CREATE TABLE applications (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  candidate_user_id UUID NOT NULL,
  vacancy_id        UUID NOT NULL,
  company_id        UUID NOT NULL,
  cover_letter      TEXT,
  status            TEXT NOT NULL DEFAULT 'sent'
                      CHECK (status IN ('sent','viewed','accepted','rejected','withdrawn')),
  status_updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_app_candidate ON applications(candidate_user_id, status, created_at DESC);
CREATE INDEX idx_app_vacancy   ON applications(vacancy_id, status, created_at DESC);

CREATE TABLE interaction_status_history (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type   TEXT NOT NULL CHECK (entity_type IN ('invitation','application')),
  entity_id     UUID NOT NULL,
  from_status   TEXT,
  to_status     TEXT NOT NULL,
  actor_user_id UUID,
  comment       TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_status_history_entity ON interaction_status_history(entity_type, entity_id, created_at);

-- лог раскрытия контактов (аудит + комплаенс 152-ФЗ)
CREATE TABLE contact_reveals (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  candidate_user_id UUID NOT NULL,
  employer_user_id  UUID NOT NULL,
  reason            TEXT NOT NULL CHECK (reason IN ('invitation_accepted','application_sent')),
  entity_type       TEXT NOT NULL CHECK (entity_type IN ('invitation','application')),
  entity_id         UUID NOT NULL,
  revealed_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (candidate_user_id, employer_user_id, entity_type, entity_id)
);