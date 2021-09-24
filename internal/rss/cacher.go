package rss

import (
	"encoding/xml"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"service-rss/internal/config"
	"service-rss/internal/database"
)

type rssWithID struct {
	id  int64
	rss *database.Rss
}

type Cacher struct {
	db           database.Database
	aggregator   Aggregator
	rssChan      chan *rssWithID
	workersCount int
	pullPeriod   time.Duration

	// graceful shutdown helper channels
	initChan         chan interface{}
	shutdownChan     chan interface{}
	shutdownWaitChan chan interface{}
}

func NewCacher(cfg *config.Config, db database.Database, aggregator Aggregator) *Cacher {
	return &Cacher{
		db:               db,
		aggregator:       aggregator,
		rssChan:          make(chan *rssWithID, cfg.CacherWorkersCount),
		workersCount:     cfg.CacherWorkersCount,
		pullPeriod:       cfg.CacherPullPeriod,
		initChan:         make(chan interface{}, cfg.CacherWorkersCount+1),
		shutdownChan:     make(chan interface{}),
		shutdownWaitChan: make(chan interface{}),
	}
}

func (c *Cacher) Start() {
	defer close(c.shutdownWaitChan)

	ticker := time.NewTicker(c.pullPeriod)

	// push tasks
	go func() {
		defer close(c.rssChan)

		first := make(chan interface{}, 1)
		defer close(first)
		first <- nil

		for {
			select {
			// pushTasks at the start
			case <-first:
				c.initChan <- nil
				c.pushTasks()
			case <-c.shutdownChan:
				return
			case <-ticker.C:
				c.pushTasks()
			}
		}
	}()

	// process tasks
	wg := sync.WaitGroup{}
	for i := 0; i < c.workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			first := make(chan interface{}, 1)
			defer close(first)
			first <- nil

			select {
			case <-first:
				c.initChan <- nil
			case <-c.shutdownChan:
				return
			case rss := <-c.rssChan:
				c.processTask(rss)
			}
		}()
	}

	wg.Wait()
}

func (c *Cacher) Shutdown() {
	// wait for task pusher + workers start
	for i := 0; i < c.workersCount+1; i++ {
		<-c.initChan
	}
	close(c.shutdownChan)
	<-c.shutdownWaitChan
}

func (c *Cacher) pushTasks() {
	rssMap, err := c.db.GetItemsToCache()
	if err != nil {
		log.WithError(err).Error("failed to get items to cache")
		return
	}

	for id, rss := range rssMap {
		c.rssChan <- &rssWithID{
			id:  id,
			rss: rss,
		}
	}
}

func (c *Cacher) processTask(rss *rssWithID) {
	rssFeed := c.aggregator.Aggregate(rss.rss)

	rssFeedRaw, err := xml.Marshal(rssFeed)
	if err != nil {
		log.WithError(err).Error("failed to serialize rss feed")
		return
	}

	validUntil := time.Now().Add(time.Duration(rssFeed.Channel.Ttl) * time.Minute)

	err = c.db.SaveCache(rss.id, string(rssFeedRaw), validUntil)
	if err != nil {
		log.WithError(err).Error("failed to save cached rss feed")
	}
}
