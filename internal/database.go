package kuura

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

type DatabaseConfig struct {
	DSN               string
	MaxConnections    int32
	MinConnections    int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectionTimeout time.Duration
}

func DefaultPostgresConfig() DatabaseConfig {
	return DatabaseConfig{
		MaxConnections:    4,
		MinConnections:    1,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 15 * time.Second,
		ConnectionTimeout: 5 * time.Second,
	}
}

func ProvidePGXConfig(logger *slog.Logger, dsn string, config ...DatabaseConfig) *pgxpool.Config {
	dbConfig := DefaultPostgresConfig()
	if len(config) > 0 {
		dbConfig = config[0]
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("Failed to parse database connection config",
			slog.String("error", err.Error()),
			slog.String("dsn", dsn),
		)
		return nil
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	poolConfig.MaxConns = dbConfig.MaxConnections
	poolConfig.MinConns = dbConfig.MinConnections
	poolConfig.MaxConnLifetime = dbConfig.MaxConnLifetime
	poolConfig.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = dbConfig.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = dbConfig.ConnectionTimeout

	poolConfig.BeforeClose = func(c *pgx.Conn) {
		logger.Debug("Closing PostgreSQL connection pool", slog.Uint64("conn_pid", uint64(c.PgConn().PID())))
	}
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		logger.Debug("Opened PostgreSQL connection",
			slog.Uint64("conn_pid", uint64(conn.PgConn().PID())))
		return nil
	}

	return poolConfig
}

type DatabaseManager struct {
	pool     *pgxpool.Pool
	sqlDB    *sql.DB
	logger   *slog.Logger
	poolOnce sync.Once
	sqlOnce  sync.Once
}

func NewDatabaseManager(logger *slog.Logger) *DatabaseManager {
	return &DatabaseManager{
		logger: logger,
	}
}

func (dm *DatabaseManager) Connect(dsn string, config ...DatabaseConfig) (*pgxpool.Pool, error) {
	var err error
	dm.poolOnce.Do(func() {
		poolConfig := ProvidePGXConfig(dm.logger, dsn, config...)
		if poolConfig == nil {
			err = errors.New("failed to create database configuration")
			return
		}

		ctx := context.Background()
		dm.pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			dm.logger.Error("Failed to create database pool",
				slog.String("error", err.Error()),
			)
			return
		}

		if err = dm.validateConnection(ctx); err != nil {
			dm.pool.Close()
			dm.pool = nil
		}
	})

	return dm.pool, err
}

func (dm *DatabaseManager) validateConnection(ctx context.Context) error {
	conn, err := dm.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	dm.logger.Info("Successfully connected to PostgreSQL")
	return nil
}

func (dm *DatabaseManager) SQLDatabase() *sql.DB {
	dm.sqlOnce.Do(func() {
		if dm.pool != nil {
			dm.sqlDB = stdlib.OpenDBFromPool(dm.pool)
		}
	})
	return dm.sqlDB
}

func (dm *DatabaseManager) CheckMigrations(
	source migrate.MigrationSource,
) (bool, error) {
	sqlDB := dm.SQLDatabase()
	if sqlDB == nil {
		return false, errors.New("database connection not established")
	}

	dm.logger.Info("Checking current migration status")

	migrations, err := source.FindMigrations()
	if err != nil {
		return false, fmt.Errorf("failed to find migrations: %w", err)
	}
	dm.logger.Info("Found embedded migrations",
		slog.Int("count", len(migrations)),
	)

	records, err := migrate.GetMigrationRecords(sqlDB, "postgres")
	if err != nil {
		return false, fmt.Errorf("failed to get migration records: %w", err)
	}
	dm.logger.Info("Got migration records from database",
		slog.Int("count", len(records)),
	)

	return len(records) != len(migrations), nil
}

func (dm *DatabaseManager) ApplyMigrations(
	source migrate.MigrationSource,
) error {
	sqlDB := dm.SQLDatabase()
	if sqlDB == nil {
		return errors.New("database connection not established")
	}

	dm.logger.Info("Applying migrations")

	amount, err := migrate.Exec(sqlDB, "postgres", source, migrate.Up)
	if err != nil {
		return fmt.Errorf("migration execution failed: %w", err)
	}

	dm.logger.Info("Applied migrations",
		slog.Int("amount", amount),
	)
	return nil
}
