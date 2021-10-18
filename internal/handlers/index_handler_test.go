package handlers

import (
	"errors"
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"service-rss/internal/auth"
	"service-rss/internal/database"
)

const (
	htmlTemplateString = "{{.Email}} {{range .RssFeeds}}{{.Name}}{{end}}"
)

func TestIndexHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	htmlTemplate, err := template.New("webpage").Parse(htmlTemplateString)
	assert.NoError(t, err)

	t.Run("auth error", func(t *testing.T) {
		db := database.NewMockDatabase(ctrl)
		db.EXPECT().GetRssForIndex().Return([]*database.Rss{
			{
				Email: "example@gmail.com",
				Name:  "name",
				Sources: []string{
					"http://google.com",
				},
			},
		}, nil)

		authHandler := auth.NewMockHandler(ctrl)
		authHandler.EXPECT().GetEmail(gomock.Any(), gomock.Any()).Return("", errors.New("err"))

		handler := &indexHandler{
			db:           db,
			authHandler:  authHandler,
			htmlTemplate: htmlTemplate,
		}

		req := httptest.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Contains(t, rr.Body.String(), " name")
	})

	t.Run("auth ok", func(t *testing.T) {
		db := database.NewMockDatabase(ctrl)
		db.EXPECT().GetRssForIndex().Return([]*database.Rss{
			{
				Email: "example@gmail.com",
				Name:  "name",
				Sources: []string{
					"http://google.com",
				},
			},
		}, nil)

		authHandler := auth.NewMockHandler(ctrl)
		authHandler.EXPECT().GetEmail(gomock.Any(), gomock.Any()).Return("example@gmail.com", nil)

		handler := &indexHandler{
			db:           db,
			authHandler:  authHandler,
			htmlTemplate: htmlTemplate,
		}

		req := httptest.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Contains(t, rr.Body.String(), "example@gmail.com name")
	})

	t.Run("db fail", func(t *testing.T) {
		db := database.NewMockDatabase(ctrl)
		db.EXPECT().GetRssForIndex().Return(nil, errors.New("error"))

		authHandler := auth.NewMockHandler(ctrl)
		authHandler.EXPECT().GetEmail(gomock.Any(), gomock.Any()).Return("example@gmail.com", nil)

		handler := &indexHandler{
			db:           db,
			authHandler:  authHandler,
			htmlTemplate: htmlTemplate,
		}

		req := httptest.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, 500, rr.Code)
		assert.Contains(t, rr.Body.String(), "{\"error\":\"failed to get feeds\",\"value\":\"error\"}")
	})
}
