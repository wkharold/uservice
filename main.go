// Package main provides a simple web server to test out some ideas
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mholt/caddy/middleware"
	"github.com/mholt/caddy/middleware/headers"
	reqlog "github.com/mholt/caddy/middleware/log"
)

type greeter struct{}

func (greeter) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Fprintf(w, "Hello World!")
	return http.StatusOK, nil
}

func main() {
	rule := reqlog.Rule{
		PathScope: "/",
		Format:    reqlog.DefaultLogFormat,
		Log:       log.New(os.Stdout, "", 0),
	}

	var g greeter
	headers := headers.Headers{
		Rules: []headers.Rule{
			{Path: "/", Headers: []headers.Header{
				{Name: "X-Middleware", Value: "Caddy"},
				{Name: "X-Honestdollar-Seed", Value: "42"},
			}},
		},
		Next: middleware.Handler(g)}
	logger := reqlog.Logger{Rules: []reqlog.Rule{rule}, Next: middleware.Handler(headers)}

	http.Handle("/", func(hf middleware.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if status, err := hf.ServeHTTP(w, r); err != nil {
				http.Error(w, err.Error(), status)
			}
		}
	}(logger))
	http.ListenAndServe(":8888", nil)
}
