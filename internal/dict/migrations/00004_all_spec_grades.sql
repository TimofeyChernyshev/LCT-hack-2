-- +goose Up
INSERT INTO categories (specialization_id, grade_id, slug)
SELECT s.id, g.id, s.code || '_' || g.code
FROM specializations s
CROSS JOIN grades g
WHERE NOT EXISTS (
  SELECT 1 FROM categories c
  WHERE c.specialization_id = s.id AND c.grade_id = g.id
);
