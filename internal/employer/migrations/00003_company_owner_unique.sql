-- +goose Up
DROP INDEX IF EXISTS idx_companies_owner;
CREATE UNIQUE INDEX idx_companies_owner ON companies(owner_user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_companies_owner;
CREATE INDEX idx_companies_owner ON companies(owner_user_id);
