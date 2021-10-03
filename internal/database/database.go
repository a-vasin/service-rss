//go:generate mockgen -package ${GOPACKAGE} -destination mock_database.go -source database.go
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"

	"service-rss/internal/config"
)

var (
	lockTimeout = 30 * time.Minute
)

type Rss struct {
	Name    string
	Sources []string
}

type Database interface {
	Shutdown() error
	CreateRss(*Rss) error
	GetItemsToCache(batchSize int) (map[int64]*Rss, error)
	SaveCache(id int64, rssFeed string, validUntil time.Time) error
}

type database struct {
	db        *sql.DB
	serviceID string
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

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	serviceID := fmt.Sprintf("%s-%d", hostname, os.Getppid())

	return &database{
		db:        db,
		serviceID: serviceID,
	}, nil
}

func (db *database) Shutdown() error {
	return db.db.Close()
}

func (db *database) CreateRss(rss *Rss) error {
	if rss == nil {
		return errors.New("empty rss")
	}

	query := "INSERT INTO rss (name, sources, cached_valid_until) VALUES ($1, $2, $3)"
	_, err := db.db.Exec(query, rss.Name, pq.Array(rss.Sources), time.Unix(0, 0))
	if err != nil {
		return err
	}
	return nil
}

func (db *database) GetItemsToCache(batchSize int) (map[int64]*Rss, error) {
	now := time.Now()

	ids, err := db.getNotLockedItems(batchSize, now)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return map[int64]*Rss{}, nil
	}

	err = db.lockItems(ids, now)
	if err != nil {
		return nil, err
	}

	items, err := db.getLockedItems(ids)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (db *database) SaveCache(id int64, rssFeed string, validUntil time.Time) error {
	query := "UPDATE rss SET is_locked=FALSE, cached_rss=$1, cached_valid_until=$2 WHERE id=$3"
	_, err := db.db.Exec(query, rssFeed, validUntil, id)
	if err != nil {
		return err
	}
	return nil
}

func (db *database) getNotLockedItems(batchSize int, now time.Time) ([]int64, error) {
	query := "SELECT id FROM rss WHERE (not is_locked or locked_time < $1) and cached_valid_until < $2 LIMIT $3"
	rows, err := db.db.Query(query, now.Add(-lockTimeout), now, batchSize)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (db *database) lockItems(ids []int64, now time.Time) error {
	query := "UPDATE rss SET is_locked=TRUE, locked_by=$1, locked_time=$2 WHERE (not is_locked or locked_time < $3) and id=any($4)"
	_, err := db.db.Exec(query, db.serviceID, now, now.Add(-lockTimeout), pq.Array(ids))
	if err != nil {
		return err
	}
	return nil
}

func (db *database) getLockedItems(ids []int64) (map[int64]*Rss, error) {
	query := "SELECT id, name, sources FROM rss WHERE is_locked and locked_by=$1 and id=any($2)"
	rows, err := db.db.Query(query, db.serviceID, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make(map[int64]*Rss, len(ids))
	for rows.Next() {
		var id int64
		item := &Rss{}
		if err = rows.Scan(&id, &item.Name, pq.Array(&item.Sources)); err != nil {
			return nil, err
		}
		items[id] = item
	}

	return items, nil
}
