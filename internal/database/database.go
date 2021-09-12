//go:generate mockgen -package ${GOPACKAGE} -destination mock_database.go -source database.go
package database

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"service-rss/internal/config"
)

type Database interface {
	Shutdown() error
	CreateRss(name string, sources []string) error
}

type database struct {
	db *sql.DB
}

func New(cfg *config.Config) (Database, error) {
	settings := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbName, cfg.DbUser, cfg.DbPassword,
	)
	if !cfg.DbEnableSsl {
		settings = fmt.Sprintf("%s sslmode=disable", settings)
	}

	db, err := sql.Open("postgres", settings)
	if err != nil {
		return nil, err
	}
	return &database{
		db: db,
	}, nil
}

func (db *database) Shutdown() error {
	return db.db.Close()
}

func (db *database) CreateRss(name string, sources []string) error {
	sqlStatement := "INSERT INTO rss (name, sources) VALUES ($1, $2)"
	_, err := db.db.Exec(sqlStatement, name, pq.Array(sources))
	if err != nil {
		return err
	}
	return nil
}
