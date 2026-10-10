-- +goose Up
CREATE TABLE companies (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id    UUID NOT NULL,
  name             TEXT NOT NULL,
  description      TEXT,
  industry         TEXT,
  website          TEXT,
  size             TEXT,
  contact_person   TEXT,
  contact_email    TEXT,
  contact_phone    TEXT,
  contact_telegram TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_companies_owner ON companies(owner_user_id);
CREATE UNIQUE INDEX idx_companies_owner_unique ON companies(owner_user_id);

CREATE TABLE employer_needs (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  title                TEXT NOT NULL,
  description          TEXT NOT NULL,
  category_id          UUID,
  specialization_id    UUID,
  grade_id             UUID,
  stack                UUID[] NOT NULL DEFAULT '{}',
  salary_min           INT NOT NULL,
  salary_max           INT NOT NULL,
  currency             TEXT NOT NULL DEFAULT 'RUB',
  work_format          TEXT CHECK (work_format IN ('remote','hybrid','office')),
  location             TEXT,
  min_experience_years NUMERIC(3,1),
  status               TEXT NOT NULL DEFAULT 'draft'
                         CHECK (status IN ('draft','active','closed')),
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (salary_min <= salary_max)
);
CREATE INDEX idx_needs_company ON employer_needs(company_id);
CREATE INDEX idx_needs_status ON employer_needs(status);
CREATE INDEX idx_needs_stack ON employer_needs USING GIN (stack);
CREATE INDEX idx_needs_category ON employer_needs(category_id);

CREATE TABLE vacancies (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  need_id           UUID REFERENCES employer_needs(id) ON DELETE SET NULL,
  title             TEXT NOT NULL,
  description       TEXT NOT NULL,
  category_id       UUID,
  specialization_id UUID,
  grade_id          UUID,
  stack             UUID[] NOT NULL DEFAULT '{}',
  salary_min        INT NOT NULL,
  salary_max        INT NOT NULL,
  currency          TEXT NOT NULL DEFAULT 'RUB',
  work_format       TEXT CHECK (work_format IN ('remote','hybrid','office')),
  location          TEXT,
  status            TEXT NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft','published','archived')),
  published_at      TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (salary_min <= salary_max)
);
CREATE INDEX idx_vacancies_company ON vacancies(company_id);
CREATE INDEX idx_vacancies_status_pub ON vacancies(status, published_at DESC);
CREATE INDEX idx_vacancies_stack ON vacancies USING GIN (stack);
CREATE INDEX idx_vacancies_category ON vacancies(category_id);

-- +goose Down
DROP TABLE vacancies;
DROP TABLE employer_needs;
DROP TABLE companies;