package main

import (
	"context"
	"log"
	"time"

	"simplejourney/helpdesk/backend/internal/config"
	"simplejourney/helpdesk/backend/internal/repository"
	"simplejourney/helpdesk/backend/internal/seeder"
	"simplejourney/helpdesk/backend/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi gagal dimuat: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("koneksi database gagal: %v", err)
	}
	defer pool.Close()

	if err := migrations.Run(ctx, pool); err != nil {
		log.Fatalf("migrasi skema gagal: %v", err)
	}

	if err := seeder.Seed(ctx, pool); err != nil {
		log.Fatalf("seeder gagal dijalankan: %v", err)
	}

	log.Println("seeder data awal berhasil dieksekusi")
}
