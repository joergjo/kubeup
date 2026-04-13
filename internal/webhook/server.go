package webhook

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// NewServer creates a new http.Server with the given handler, port and path.
// The handler is expected to provide the webhook functionality.
func NewServer(handler http.Handler, opts ...Options) (*http.Server, error) {
	var options options
	var err error
	for _, opt := range opts {
		if err = opt(&options); err != nil {
			return nil, fmt.Errorf("invalid options: %w", err)

		}
	}

	mux := http.NewServeMux()
	switch {
	case options.entraID != nil:
		v, err := newValidator(options.entraID.issuerURL, options.entraID.appID)
		if err != nil {
			return nil, fmt.Errorf("creating JWT validator: %w", err)
		}
		mw, err := newEntraIDMiddleware(v)
		if err != nil {
			return nil, fmt.Errorf("creating JWT middleware: %w", err)
		}
		mux.Handle(options.path, mw(handler))
		slog.Info("using EntraID middleware")
	case options.clientSecret != nil:
		mw := newClientSecretMiddleware(options.clientSecret.secret1, options.clientSecret.secret2)
		mux.Handle(options.path, mw(handler))
		slog.Info("using client secret middleware")
	default:
		mux.Handle(options.path, handler)
		slog.Warn("no authorization middleware configured, webhook is unprotected")
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	s := http.Server{
		Addr:              fmt.Sprintf(":%d", options.port),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 15, // 32KB
		Handler:           mux,
	}
	return &s, nil
}
