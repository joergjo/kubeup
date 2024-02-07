package webhook

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

// New creates a new http.Server with the given handler, port and path.
// The handler is expected to provide the webhook functionality.
func New(h http.Handler, port int, path string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(path, h)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
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
	log.Warn().Msgf("received signal %v, shutting down", sig)

	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := s.Shutdown(childCtx); err != nil {
		log.Error().Err(err).Msg("shutting down server")
	}
	close(done)
}
