package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"

	"service-rss/internal/database"
	"service-rss/internal/dto"
)

type rssCreateHandler struct {
	db     database.Database
	schema *gojsonschema.Schema
}

func NewRssCreateHandler(db database.Database, schema *gojsonschema.Schema) http.Handler {
	return &rssCreateHandler{
		db:     db,
		schema: schema,
	}
}

func (h *rssCreateHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeBadRequest(writer, "failed to read request body", "")
		return
	}

	loader := gojsonschema.NewBytesLoader(bodyBytes)
	result, err := h.schema.Validate(loader)
	if err != nil {
		writeBadRequest(writer, "failed to validate input", string(bodyBytes))
		return
	}

	if !result.Valid() {
		response := []string{"input validation failed:"}
		for _, desc := range result.Errors() {
			response = append(response, fmt.Sprintf("- %s", desc))
		}
		errors := strings.Join(response, "\n")

		writeBadRequest(writer, "input validation failed", errors)
		return
	}

	in := &dto.RssCreateIn{}
	err = json.Unmarshal(bodyBytes, &in)
	if err != nil {
		writeBadRequest(writer, "failed to unmarshal input", string(bodyBytes))
		return
	}

	wrongUrls := make([]string, 0, len(in.Sources))
	for _, rawUrl := range in.Sources {
		isUrl := govalidator.IsURL(rawUrl)
		if !isUrl {
			wrongUrls = append(wrongUrls, rawUrl)
		}
	}

	if len(wrongUrls) > 0 {
		urls := strings.Join(wrongUrls, "\n")
		writeBadRequest(writer, "found malformed input source urls", urls)
		return
	}

	rss := &database.Rss{
		Name:    in.Name,
		Sources: in.Sources,
	}
	err = h.db.CreateRss(rss)
	if err != nil {
		writeInternalError(writer, "failed to create rss", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

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
	}

	writeErrorResponse(writer, http.StatusInternalServerError, resp)
}
