package rss

import (
	"encoding/xml"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"service-rss/internal/config"
	"service-rss/internal/database"
	"service-rss/internal/dto"
	"service-rss/internal/safe"
)

type Cacher struct {
	db           database.Database
	aggregator   Aggregator
	rssChan      chan *database.Rss
	workersCount int
	pullPeriod   time.Duration
	batchSize    int

	// graceful shutdown helper-channels
	shutdownChan     chan interface{}
	shutdownWaitChan chan interface{}
}

func NewCacher(cfg *config.Config, db database.Database, aggregator Aggregator) *Cacher {
	return &Cacher{
		db:           db,
		aggregator:   aggregator,
		rssChan:      make(chan *database.Rss, cfg.CacherWorkersCount),
		workersCount: cfg.CacherWorkersCount,
		pullPeriod:   cfg.CacherPullPeriod,
		batchSize:    cfg.CacherBatchSize,

		shutdownChan:     make(chan interface{}),
		shutdownWaitChan: make(chan interface{}),
	}
}

func (c *Cacher) Start() {
	defer close(c.shutdownWaitChan)

	ticker := time.NewTicker(c.pullPeriod)

	// push tasks
	go safe.Do(func() {
		defer close(c.rssChan)

		first := make(chan interface{}, 1)
		defer close(first)
		first <- nil

		for {
			select {
			// pushTasks at the start
			case <-first:
				c.pushTasks()
			case <-c.shutdownChan:
				return
			case <-ticker.C:
				c.pushTasks()
			}
		}
	})

	// process tasks
	wg := sync.WaitGroup{}
	for i := 0; i < c.workersCount; i++ {
		wg.Add(1)
		go safe.Do(func() {
			defer wg.Done()

			for {
				select {
				case <-c.shutdownChan:
					return
				case rss := <-c.rssChan:
					c.processTask(rss)
				}
			}
		})
	}

	wg.Wait()
}

func (c *Cacher) Shutdown() {
	close(c.shutdownChan)
	<-c.shutdownWaitChan
}

func (c *Cacher) pushTasks() {
	rssSlice, err := c.db.GetItemsToCache(c.batchSize)
	if err != nil {
		log.WithError(err).Error("failed to get items to cache")
		return
	}

	for _, rss := range rssSlice {
		c.rssChan <- rss
	}
}

func (c *Cacher) processTask(rss *database.Rss) {
	rssFeed := c.aggregator.Aggregate(rss)

	rssFeedRaw, err := xml.Marshal(rssFeed)
	if err != nil {
		log.WithError(err).Error("failed to serialize rss feed")
		return
	}

	validUntil := GetValidUntil(rssFeed)

	err = c.db.SaveCachedRss(rss.ID, string(rssFeedRaw), validUntil)
	if err != nil {
		log.WithError(err).Error("failed to save cached rss feed")
	}

	log.WithField("name", rss.Name).WithField("email", rss.Email).Info("rss was processed")
}

func GetValidUntil(rssFeed *dto.RssFeed) time.Time {
	return time.Now().Add(time.Duration(rssFeed.Channel.Ttl) * time.Minute)
}
