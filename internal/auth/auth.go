//go:generate mockgen -package ${GOPACKAGE} -destination mock_auth.go -source auth.go
package auth

import "net/http"

type Handler interface {
	Login(w http.ResponseWriter, r *http.Request)
	GetEmail(w http.ResponseWriter, r *http.Request) (string, error)
}
