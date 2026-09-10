// Package httpapi provides the HTTP transport for the application.
package httpapi

import (
	"fmt"
	"net/http"
)

// NewHandler creates the application HTTP handler.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}
