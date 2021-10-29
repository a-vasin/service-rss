package rss

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"service-rss/internal/config"
	"service-rss/internal/database"
)

func TestCacher_Shutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := NewMockFetcher(ctrl)
	f.EXPECT().Fetch(gomock.Any()).AnyTimes().Return(nil, nil)

	a := NewTestAggregator(f)

	db := database.NewMockDatabase(ctrl)
	db.EXPECT().GetItemsToCache(gomock.Any()).AnyTimes().Return(nil, nil)

	cfg := &config.Config{
		CacherPullPeriod:   30 * time.Second,
		CacherWorkersCount: 4,
	}

	h := NewCacher(cfg, db, a)

	timeout := time.After(1 * time.Second)
	done := make(chan bool)
	go func() {
		go h.Start()
		h.Shutdown()

		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("test didn't finish in time")
	case <-done:
	}
}

func TestCacher_PushTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := database.NewMockDatabase(ctrl)
	db.EXPECT().GetItemsToCache(gomock.Any()).AnyTimes().Return([]*database.Rss{
		{
			ID:   1,
			Name: "first",
		},
		{
			ID:   2,
			Name: "second",
		},
	}, nil)

	h := &Cacher{
		db:      db,
		rssChan: make(chan *database.Rss, 2),
	}

	timeout := time.After(1 * time.Second)
	resultChan := make(chan []*database.Rss)
	go func() {
		h.pushTasks()

		result := make([]*database.Rss, 0, 2)
		for i := 0; i < 2; i++ {
			rss := <-h.rssChan
			result = append(result, rss)
		}

		resultChan <- result
	}()

	expected := []*database.Rss{
		{
			ID:   1,
			Name: "first",
		},
		{
			ID:   2,
			Name: "second",
		},
	}

	select {
	case <-timeout:
		t.Fatal("test didn't finish in time")
	case actual := <-resultChan:
		sort.Slice(actual, func(i, j int) bool {
			return actual[i].ID < actual[j].ID
		})
		assert.Equal(t, expected, actual)
	}
}

func TestCacher_ProcessTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := NewMockFetcher(ctrl)

	a := NewTestAggregator(f)

	db := database.NewMockDatabase(ctrl)
	db.EXPECT().SaveCachedRss(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(id int64, rssFeed string, validUntil time.Time) {
		assert.Equal(t, int64(1), id)
		assert.True(t, strings.HasPrefix(rssFeed, "<rss><channel><title>RSS Aggregator</title><link></link><description>Aggregated feed from different rss sources.</description><lastBuildDate>"))
		assert.True(t, strings.HasSuffix(rssFeed, "</lastBuildDate><ttl>5</ttl></channel></rss>"))
		assert.True(t, time.Now().Before(validUntil))
	})

	h := &Cacher{
		db:         db,
		aggregator: a,
	}

	rss := &database.Rss{
		ID:   1,
		Name: "name",
	}
	h.processTask(rss)
}
