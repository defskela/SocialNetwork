package migrator

import (
	"errors"
	"fmt"
	"log"
	"time"

	"social-network/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MustRun(cfg config.Postgres, migrationsPath string) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	var m *migrate.Migrate
	var err error

	for i := 0; i < 10; i++ {
		m, err = migrate.New("file://"+migrationsPath, dsn)
		if err == nil {
			break
		}
		log.Printf("Migrator: failed to connect to database: %v. Retrying in 2s...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Migrator: failed to initialize after retries: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Migrator: no changes to apply")
			return
		}
		log.Fatalf("Migrator: failed to apply migrations: %v", err)
	}

	log.Println("Migrator: migrations applied successfully")
}
