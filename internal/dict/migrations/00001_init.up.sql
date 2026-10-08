CREATE TABLE industries (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       TEXT UNIQUE NOT NULL,
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE specializations (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  industry_id   UUID NOT NULL REFERENCES industries(id),
  code          TEXT UNIQUE NOT NULL,
  name          TEXT NOT NULL,
  description   TEXT,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_specializations_industry ON specializations(industry_id);

CREATE TABLE grades (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       TEXT UNIQUE NOT NULL,
  name       TEXT NOT NULL,
  rank       SMALLINT NOT NULL UNIQUE, -- 1..N, порядок
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  specialization_id UUID NOT NULL REFERENCES specializations(id),
  grade_id          UUID NOT NULL REFERENCES grades(id),
  slug              TEXT UNIQUE NOT NULL,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (specialization_id, grade_id)
);
CREATE INDEX idx_categories_grade ON categories(grade_id);

CREATE TABLE technologies (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       TEXT UNIQUE NOT NULL,
  name       TEXT NOT NULL,
  category   TEXT, -- language, db, framework, cloud, ...
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);