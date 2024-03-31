package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	secP = "secret"
)

type ServerOptions struct {
	Path    string
	Port    int
	Secret1 string
	Secret2 string
}

// NewServer creates a new http.Server with the given handler, port and path.
// The handler is expected to provide the webhook functionality.
func NewServer(h http.Handler, opts ServerOptions) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(opts.Path, protectWithClientSecret(h, opts.Secret1, opts.Secret2))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	s := http.Server{
		Addr:         fmt.Sprintf(":%d", opts.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	return &s
}

// Shutdown gracefully shuts down the server when a SIGINT or SIGTERM is received.
func Shutdown(ctx context.Context, s *http.Server, done chan<- struct{}, timeout time.Duration) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigch
	slog.Warn("received signal, shutting down", "signal", sig.String())

	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := s.Shutdown(childCtx); err != nil {
		slog.Error("shutting down server", "error", err)
	}
	close(done)
}

// TODO: add test
func newClientSecretMiddleware(sec1, sec2 string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := r.URL.Query().Get(secP)
			if secret != sec1 && secret != sec2 {
				slog.Warn("received request with invalid secret")
				http.Error(w, "invalid secret", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func protectWithClientSecret(h http.Handler, sec1, sec2 string) http.Handler {
	if sec1 == "" || sec2 == "" {
		return h
	}
	return newClientSecretMiddleware(sec1, sec2)(h)
}
