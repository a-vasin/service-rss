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
	ID      int64
	Email   string
	Name    string
	Sources []string
}

type RssCached struct {
	Rss
	RssFeed string
}

type Database interface {
	Shutdown() error
	CreateRss(*Rss) error
	GetItemsToCache(batchSize int) ([]*Rss, error)
	SaveCachedRss(id int64, rssFeed string, validUntil time.Time) error
	GetCachedRss(email string, name string) (*RssCached, error)
	GetRssForIndex() ([]*Rss, error)
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

	query := "INSERT INTO rss (email, name, sources, cached_valid_until) VALUES ($1, $2, $3, $4)"
	_, err := db.db.Exec(query, rss.Email, rss.Name, pq.Array(rss.Sources), time.Unix(0, 0))
	if err != nil {
		return err
	}
	return nil
}

func (db *database) GetItemsToCache(batchSize int) ([]*Rss, error) {
	now := time.Now()

	ids, err := db.getNotLockedItems(batchSize, now)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*Rss{}, nil
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

func (db *database) SaveCachedRss(id int64, rssFeed string, validUntil time.Time) error {
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

func (db *database) getLockedItems(ids []int64) ([]*Rss, error) {
	query := "SELECT id, email, name, sources FROM rss WHERE is_locked and locked_by=$1 and id=any($2)"
	rows, err := db.db.Query(query, db.serviceID, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*Rss, 0, len(ids))
	for rows.Next() {
		item := &Rss{}
		if err = rows.Scan(&item.ID, &item.Email, &item.Name, pq.Array(&item.Sources)); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (db *database) GetCachedRss(email string, name string) (*RssCached, error) {
	query := "SELECT sources, cached_rss FROM rss WHERE email=$1 and name=$2"
	row := db.db.QueryRow(query, email, name)

	var rssFeed string
	var sources []string
	err := row.Scan(pq.Array(&sources), &rssFeed)
	if err != nil {
		return nil, err
	}

	return &RssCached{
		Rss: Rss{
			Email:   email,
			Name:    name,
			Sources: sources,
		},
		RssFeed: rssFeed,
	}, nil
}

func (db *database) GetRssForIndex() ([]*Rss, error) {
	query := "SELECT id, email, name, sources FROM rss ORDER BY added_time desc"
	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]*Rss, 0)
	for rows.Next() {
		item := &Rss{}
		if err = rows.Scan(&item.ID, &item.Email, &item.Name, pq.Array(&item.Sources)); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
