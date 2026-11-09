package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 000001_init_schema.sql
var initSchemaSQL string

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, initSchemaSQL); err != nil {
		return fmt.Errorf("execute init schema migration: %w", err)
	}
	return nil
}
