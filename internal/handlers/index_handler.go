package handlers

import (
	"fmt"
	"net/http"

	"service-rss/internal/auth"
	"service-rss/internal/database"
)

type indexHandler struct {
	db          database.Database
	authHandler auth.Handler
}

func NewIndexHandler(db database.Database, authHandler auth.Handler) http.Handler {
	return &indexHandler{
		db:          db,
		authHandler: authHandler,
	}
}

func (h *indexHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	email, err := h.authHandler.GetEmail(writer, req)
	if err != nil {
		writeBadRequest(writer, "failed to get email", err.Error())
		return
	}

	fmt.Println(email)

	writer.WriteHeader(http.StatusOK)
}
