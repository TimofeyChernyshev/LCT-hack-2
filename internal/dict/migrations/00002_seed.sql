-- +goose Up
INSERT INTO industries (code, name) VALUES
  ('it', 'Информационные технологии');

INSERT INTO specializations (industry_id, code, name) VALUES
  ((SELECT id FROM industries WHERE code='it'), 'backend',  'Backend-разработчик'),
  ((SELECT id FROM industries WHERE code='it'), 'frontend', 'Frontend-разработчик'),
  ((SELECT id FROM industries WHERE code='it'), 'mobile',   'Мобильный разработчик'),
  ((SELECT id FROM industries WHERE code='it'), 'devops',   'DevOps-инженер'),
  ((SELECT id FROM industries WHERE code='it'), 'data',     'Data-инженер/аналитик'),
  ((SELECT id FROM industries WHERE code='it'), 'qa',       'QA-инженер');

INSERT INTO grades (code, name, rank) VALUES
  ('intern',      'Intern',  1),
  ('junior',      'Junior',  2),
  ('junior_plus', 'Junior+', 3),
  ('middle',      'Middle',  4),
  ('middle_plus', 'Middle+', 5),
  ('senior',      'Senior',  6),
  ('lead',        'Lead',    7);

INSERT INTO categories (specialization_id, grade_id, slug)
SELECT s.id, g.id, s.code || '_' || g.code
FROM specializations s, grades g
WHERE s.code IN ('backend','frontend');

INSERT INTO technologies (code, name, category) VALUES
  ('go','Go','language'), ('python','Python','language'),
  ('java','Java','language'), ('kotlin','Kotlin','language'),
  ('ts','TypeScript','language'), ('js','JavaScript','language'),
  ('postgres','PostgreSQL','db'), ('mysql','MySQL','db'),
  ('mongo','MongoDB','db'), ('redis','Redis','db'),
  ('docker','Docker','infra'), ('k8s','Kubernetes','infra'),
  ('react','React','frontend'), ('vue','Vue','frontend'),
  ('nginx','Nginx','infra');
