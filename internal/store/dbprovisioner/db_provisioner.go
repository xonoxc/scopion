package dbprovisioner

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/xonoxc/scopion/internal/infra/dockerclient"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
	"github.com/xonoxc/scopion/internal/store/migrations"
)

type Provisioner struct {
	migrator *migrations.Migrator
}

func New(migrator *migrations.Migrator) *Provisioner {
	return &Provisioner{
		migrator: migrator,
	}
}

func (p *Provisioner) Provision(dialect string, dsn string) (*sql.DB, error) {
	s, err := p.ValidateDSN(dsn)
	if err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if s == "" {
		dc, err := dockerclient.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create docker client: %w", err)
		}

		containerRes, err := dc.RunContainer(connCtx, dockerclient.ContainerSpec{
			Name:  "scopion-secondary-db",
			Image: "postgres:16",
			Env: []string{
				"POSTGRES_USER=scopion",
				"POSTGRES_PASSWORD=scopion",
				"POSTGRES_DB=scopion_secondary",
			},
			Ports: map[string]string{
				"5432/tcp": "5432",
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run postgres container: %w", err)
		}

		dsn = containerRes.ConnectDSN
	}

	db, err := p.connectDatabase(connCtx, dsn)
	if err != nil {
		return nil, err
	}

	if dialect == "postgres" {
		if err := p.MigrateSecondary(db); err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

func (p *Provisioner) connectDatabase(context context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(context); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (p *Provisioner) MigrateSecondary(dbConn *sql.DB) error {
	err := p.migrator.AquireMigrationLock(dbConn)
	if err != nil {
		return err
	}

	if err := p.migrator.Migrate(dbConn, migrateable.POSTGRES); err != nil {
		return err
	}

	err = p.migrator.ReleaseMigrationLock(dbConn)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) ValidateDSN(dsn string) (string, error) {
	if strings.TrimSpace(dsn) == "" {
		return "", nil
	}

	if _, err := pgx.ParseConfig(dsn); err != nil {
		return "", fmt.Errorf("invalid DSN: %w", err)
	}

	return dsn, nil
}
