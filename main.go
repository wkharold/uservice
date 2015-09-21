// Package main provides a simple web server to test out some ideas
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/mholt/caddy/middleware"
	"github.com/mholt/caddy/middleware/headers"
	mwlog "github.com/mholt/caddy/middleware/log"
)

type ContextHandler interface {
	ServeHTTPWithContext(context.Context, http.ResponseWriter, *http.Request) (int, error)
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request) (int, error)

func (h ContextHandlerFunc) ServeHTTPWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	return h(ctx, w, r)
}

type ContextAdapter struct {
	ctx     context.Context
	handler ContextHandler
}

func (cw ContextAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return cw.handler.ServeHTTPWithContext(cw.ctx, w, r)
}

func greeter(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	fmt.Fprintf(w, "Hello World!")
	return http.StatusOK, nil
}

type MiddlewareWrapper struct {
	handler middleware.Handler
}

func (m MiddlewareWrapper) ServeHTTPWithContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	return m.handler.ServeHTTP(w, r)
}

func main() {
	g := &ContextAdapter{
		ctx:     context.Background(),
		handler: ContextHandlerFunc(greeter),
	}

	headers := headers.Headers{
		Rules: []headers.Rule{
			{Path: "/", Headers: []headers.Header{
				{Name: "X-Honestdollar-Chanllenge", Value: "42"},
			}},
		},
		Next: middleware.Handler(g),
	}

	h := &ContextAdapter{
		ctx:     context.Background(),
		handler: MiddlewareWrapper{handler: headers},
	}

	logger := mwlog.Logger{
		Rules: []mwlog.Rule{
			{
				PathScope: "/",
				Format:    mwlog.DefaultLogFormat,
				Log:       log.New(os.Stdout, "", 0),
			},
		},
		Next: middleware.Handler(h),
	}

	l := &ContextAdapter{
		ctx:     context.Background(),
		handler: MiddlewareWrapper{handler: logger},
	}

	http.Handle("/", func(hf middleware.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if status, err := hf.ServeHTTP(w, r); err != nil {
				http.Error(w, err.Error(), status)
			}
		}
	}(l))
	http.ListenAndServe(":8888", nil)
}
