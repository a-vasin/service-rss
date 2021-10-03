package handlers

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"

	"service-rss/internal/auth"
	"service-rss/internal/database"
)

const (
	schema = "{\"type\":\"object\",\"description\":\"Inputfor/rss/create\",\"required\":[\"name\",\"sources\"],\"additionalProperties\":false,\"properties\":{\"name\":{\"type\":\"string\",\"pattern\":\"^[a-zA-Z0-9]+$\"},\"sources\":{\"type\":\"array\",\"minLength\":1,\"items\":{\"type\":\"string\",\"minLength\":1}}}}"
)

func TestRssCreateHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := database.NewMockDatabase(ctrl)
	db.EXPECT().CreateRss(&database.Rss{
		Email: "example@gmail.com",
		Name:  "exists",
		Sources: []string{
			"http://google.com",
		},
	}).Return(&pq.Error{Constraint: "rss_email_name_key"})
	db.EXPECT().CreateRss(&database.Rss{
		Email: "example@gmail.com",
		Name:  "error",
		Sources: []string{
			"http://google.com",
		},
	}).Return(errors.New("error"))
	db.EXPECT().CreateRss(&database.Rss{
		Email: "example@gmail.com",
		Name:  "ok",
		Sources: []string{
			"http://google.com",
		},
	}).Return(nil)

	loader := gojsonschema.NewStringLoader(schema)
	jsonSchema, err := gojsonschema.NewSchema(loader)
	assert.NoError(t, err)

	authHandler := auth.NewMockHandler(ctrl)
	authHandler.EXPECT().GetEmail(gomock.Any(), gomock.Any()).AnyTimes().Return("example@gmail.com", nil)

	defaultHandler := NewRssCreateHandler(db, jsonSchema, authHandler)

	t.Run("auth error", func(t *testing.T) {
		authHandler := auth.NewMockHandler(ctrl)
		authHandler.EXPECT().GetEmail(gomock.Any(), gomock.Any()).Return("", errors.New("err"))

		handler := NewRssCreateHandler(db, jsonSchema, authHandler)

		req := httptest.NewRequest("POST", "/api/rss/create", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "failed to get email")
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/rss/create", nil)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "failed to validate input")
	})

	t.Run("malformed input", func(t *testing.T) {
		body := strings.NewReader("{\"name\":\"example\"}")
		req := httptest.NewRequest("POST", "/api/rss/create", body)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "sources is required")
	})

	t.Run("malformed source", func(t *testing.T) {
		body := strings.NewReader("{\"name\":\"example\",\"sources\":[\"wrong_url\"]}")
		req := httptest.NewRequest("POST", "/api/rss/create", body)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "found malformed input source urls")
	})

	t.Run("exists", func(t *testing.T) {
		body := strings.NewReader("{\"name\":\"exists\",\"sources\":[\"http://google.com\"]}")
		req := httptest.NewRequest("POST", "/api/rss/create", body)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 400, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "rss already exists")
	})

	t.Run("error", func(t *testing.T) {
		body := strings.NewReader("{\"name\":\"error\",\"sources\":[\"http://google.com\"]}")
		req := httptest.NewRequest("POST", "/api/rss/create", body)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 500, rr.Code)
		assert.Equal(t, "application/json; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Body.String(), "failed to create rss")
	})

	t.Run("ok", func(t *testing.T) {
		body := strings.NewReader("{\"name\":\"ok\",\"sources\":[\"http://google.com\"]}")
		req := httptest.NewRequest("POST", "/api/rss/create", body)
		rr := httptest.NewRecorder()
		defaultHandler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, rr.Header().Get("Content-Type"), "")
		assert.Equal(t, rr.Body.String(), "")
	})
}
