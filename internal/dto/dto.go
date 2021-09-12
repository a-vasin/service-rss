package dto

type ErrorResponse struct {
	Error string `json:"error"`
	Value string `json:"value,omitempty"`
}
