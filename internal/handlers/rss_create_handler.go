package handlers

import "net/http"

type rssCreateHandler struct {
}

func NewRssCreateHandler() http.Handler {
	return &rssCreateHandler{}
}

func (h *rssCreateHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
