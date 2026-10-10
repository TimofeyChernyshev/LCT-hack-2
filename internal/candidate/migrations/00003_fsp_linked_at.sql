-- +goose Up
ALTER TABLE candidate_profiles
  ADD COLUMN IF NOT EXISTS fsp_linked_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE candidate_profiles
  DROP COLUMN IF EXISTS fsp_linked_at;
