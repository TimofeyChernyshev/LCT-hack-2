-- +goose Up
ALTER TABLE invitations ADD COLUMN IF NOT EXISTS company_name TEXT;

-- +goose Down
ALTER TABLE invitations DROP COLUMN IF EXISTS company_name;
