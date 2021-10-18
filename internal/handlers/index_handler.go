package handlers

import (
	"html/template"
	"net/http"
	"os"

	"service-rss/internal/auth"
	"service-rss/internal/database"
)

type templateData struct {
	Email    string
	RssFeeds []*database.Rss
}

type indexHandler struct {
	db           database.Database
	authHandler  auth.Handler
	htmlTemplate *template.Template
}

func NewIndexHandler(db database.Database, authHandler auth.Handler) (http.Handler, error) {
	rawTemplate, err := os.ReadFile("html/index.html")
	if err != nil {
		return nil, err
	}

	htmlTemplate, err := template.New("webpage").Parse(string(rawTemplate))
	if err != nil {
		return nil, err
	}

	return &indexHandler{
		db:           db,
		authHandler:  authHandler,
		htmlTemplate: htmlTemplate,
	}, nil
}

func (h *indexHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	// skip error, show page without login data
	email, _ := h.authHandler.GetEmail(writer, req)

	rssFeeds, err := h.db.GetRssForIndex()
	if err != nil {
		writeInternalError(writer, "failed to get feeds", err)
		return
	}

	data := templateData{
		Email:    email,
		RssFeeds: rssFeeds,
	}

	err = h.htmlTemplate.Execute(writer, data)
	if err != nil {
		writeInternalError(writer, "failed to fill template", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
