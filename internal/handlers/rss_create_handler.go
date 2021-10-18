package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"
	"github.com/xeipuuv/gojsonschema"

	"service-rss/internal/auth"
	"service-rss/internal/database"
	"service-rss/internal/dto"
)

type rssCreateHandler struct {
	db          database.Database
	schema      *gojsonschema.Schema
	authHandler auth.Handler
}

func NewRssCreateHandler(db database.Database, schema *gojsonschema.Schema, authHandler auth.Handler) http.Handler {
	return &rssCreateHandler{
		db:          db,
		schema:      schema,
		authHandler: authHandler,
	}
}

func (h *rssCreateHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	email, err := h.authHandler.GetEmail(writer, req)
	if err != nil || len(email) == 0 {
		writeBadRequest(writer, "failed to get email", err.Error())
		return
	}

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
		Email:   email,
		Name:    in.Name,
		Sources: in.Sources,
	}
	err = h.db.CreateRss(rss)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Constraint == "rss_email_name_key" {
			writeBadRequest(writer, "rss already exists", rss.Name)
			return
		}

		writeInternalError(writer, "failed to create rss", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
