package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 000001_init_schema.sql
var initSchemaSQL string

//go:embed 000002_attachments_and_sla.sql
var attachmentsAndSlaSQL string

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, initSchemaSQL); err != nil {
		return fmt.Errorf("execute init schema migration: %w", err)
	}
	if _, err := pool.Exec(ctx, attachmentsAndSlaSQL); err != nil {
		return fmt.Errorf("execute attachments and sla migration: %w", err)
	}
	return nil
}
