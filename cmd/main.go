package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joergjo/go-samples/kubeup"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Adjust zerlog's configuration so it mirrors the CloudEvents SDK log output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.MessageFieldName = "msg"
	zerolog.TimestampFieldName = "ts"

	port := flag.Int("port", 8000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")

	flag.Parse()

	var apiKey, from, to, subject string
	var notifier kubeup.Notifier = kubeup.LogNotifier{}
	if getConfigFromEnv(&apiKey, &from, &to, &subject) {
		notifier = kubeup.NewSendGridNotifier(apiKey, from, to, subject)
	}

	h, err := kubeup.NewCloudEventHandler(context.Background(), notifier)
	if err != nil {
		log.Fatal().Err(err).Msg("Fatal error creating CloudEvent receiver")
	}

	srv := newServer(*port, *path, h)
	srvClosed := make(chan struct{})
	go shutdown(srv, srvClosed, 10*time.Second)
	log.Info().Msgf("Starting server on port %d", *port)
	err = srv.ListenAndServe()
	log.Info().Msgf("Waiting for server to shut down...")
	<-srvClosed
	log.Print(err)
}

func getConfigFromEnv(apiKey, from, to, sub *string) bool {
	var ok bool
	*apiKey, ok = os.LookupEnv("KU_SENDGRID_APIKEY")
	if !ok {
		return false
	}
	*from, ok = os.LookupEnv("KU_SENDGRID_FROM")
	if !ok {
		return false
	}
	*to, ok = os.LookupEnv("KU_SENDGRID_TO")
	if !ok {
		return false
	}
	*sub, ok = os.LookupEnv("KU_SENDGRID_SUBJECT")
	return ok
}

func newServer(port int, path string, h http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(path, h)
	return &http.Server{Addr: fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux}
}

func shutdown(srv *http.Server, srvClosed chan<- struct{}, timeout time.Duration) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigch
	log.Printf("Received signal %v, shutting down", sig)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	}
	srvClosed <- struct{}{}
}
