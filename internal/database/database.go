package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(databaseURL string) (*DB, error) {
	// Parse config with performance-optimized settings
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Connection pool optimization — Supabase pooler limita a 15 conexões
	config.MaxConns = 10                              // Max connections (abaixo do limite de 15 do pooler)
	config.MinConns = 2                              // Min idle connections
	config.MaxConnLifetime = time.Hour               // Max lifetime of a connection
	config.MaxConnIdleTime = 15 * time.Minute        // Max idle time before closing
	config.HealthCheckPeriod = 5 * time.Minute       // Period between health checks

	// Statement cache for better performance
	config.ConnConfig.RuntimeParams["statement_cache"] = "describe"

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Printf("Database connection pool configured: min=%d, max=%d, max_lifetime=%v", 
		config.MinConns, config.MaxConns, config.MaxConnLifetime)
	return &DB{Pool: pool}, nil
}

func RunMigrations(databaseURL string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// If we hit a dirty version, force it back so migrations can re-run
		if fixErr := m.Force(26); fixErr != nil {
			return fmt.Errorf("failed to run migrations: %w (force fix also failed: %v)", err, fixErr)
		}
		if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations (after force fix): %w", retryErr)
		}
	}

	log.Println("Database migrations applied successfully")
	return nil
}

func (db *DB) RunMigrations(databaseURL string) error {
	return RunMigrations(databaseURL)
}

func (db *DB) Close() {
	db.Pool.Close()
}
