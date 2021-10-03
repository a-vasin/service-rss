package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"service-rss/internal/database"
	"service-rss/internal/rss"
)

func TestRssGetHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := database.NewMockDatabase(ctrl)
	db.EXPECT().GetCachedRss(gomock.Any(), "no_rows").Return(nil, sql.ErrNoRows)
	db.EXPECT().GetCachedRss(gomock.Any(), "error").Return(nil, errors.New("error"))
	db.EXPECT().GetCachedRss(gomock.Any(), "empty").Return(&database.RssCached{}, nil)
	db.EXPECT().GetCachedRss(gomock.Any(), "ok").Return(&database.RssCached{RssFeed: "ok"}, nil)

	fetcher := rss.NewMockFetcher(ctrl)
	fetcher.EXPECT().Fetch(gomock.Any()).AnyTimes().Return(nil, nil)
	aggregator := rss.NewAggregator(fetcher)

	defaultHandler := NewRssGetHandler(db, aggregator)

	t.Run("empty email", func(t *testing.T) {
		req := createReq("", "name")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "email and name should be specified")
	})

	t.Run("empty name", func(t *testing.T) {
		req := createReq("example@gmail.com", "")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "email and name should be specified")
	})

	t.Run("no rows", func(t *testing.T) {
		req := createReq("example@gmail.com", "no_rows")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "rss feed was not foun")
	})

	t.Run("db error", func(t *testing.T) {
		req := createReq("example@gmail.com", "error")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 500, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "failed to get cached rss")
	})

	t.Run("empty cache", func(t *testing.T) {
		req := createReq("example@gmail.com", "empty")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, "application/xml", rr.Header().Get("Content-Type"))
		body := rr.Body.String()
		assert.True(t, strings.HasPrefix(body, "<rss><channel><title>RSS Aggregator</title><link></link><description>Aggregated feed from different rss sources.</description><lastBuildDate>"))
		assert.True(t, strings.HasSuffix(body, "</lastBuildDate><ttl>5</ttl></channel></rss>"))
	})

	t.Run("cache hit", func(t *testing.T) {
		req := createReq("example@gmail.com", "ok")
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, "application/xml", rr.Header().Get("Content-Type"))
		assert.Equal(t, "ok", rr.Body.String())
	})
}

func createReq(email string, name string) *http.Request {
	target := fmt.Sprintf("/%s/%s", email, name)
	req := httptest.NewRequest("GET", target, nil)

	routeContext := &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"email", "name"},
			Values: []string{email, name},
		},
	}

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeContext)
	return req.WithContext(ctx)
}
