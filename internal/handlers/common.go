package handlers

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"service-rss/internal/dto"
)

func writeErrorResponse(writer http.ResponseWriter, status int, resp *dto.ErrorResponse) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	writer.WriteHeader(status)

	response, err := json.Marshal(resp)
	if err != nil {
		log.WithError(err).Error("failed to serialize error response")
	}

	_, err = writer.Write(response)
	if err != nil {
		log.WithError(err).Warn("failed to write response")
	}
}

func writeBadRequest(writer http.ResponseWriter, responseErr string, value string) {
	log.WithField("value", value).Warn(responseErr)

	resp := &dto.ErrorResponse{
		Error: responseErr,
		Value: value,
	}

	writeErrorResponse(writer, http.StatusBadRequest, resp)
}

func writeInternalError(writer http.ResponseWriter, responseErr string, err error) {
	log.WithError(err).Error(responseErr)

	resp := &dto.ErrorResponse{
		Error: responseErr,
		Value: err.Error(),
	}

	writeErrorResponse(writer, http.StatusInternalServerError, resp)
}
