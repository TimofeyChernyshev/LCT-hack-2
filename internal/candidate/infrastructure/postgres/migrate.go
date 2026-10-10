package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations прогоняет goose up по embedded-миграциям сервиса.
// Открывает отдельное database/sql соединение, потому что goose работает
// через database/sql, а не через pgxpool.
func RunMigrations(ctx context.Context, dsn string, fsys fs.FS, logger *slog.Logger) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	logger.Info("migrations applied")
	return nil
}
