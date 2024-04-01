package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joergjo/go-samples/kubeup/internal/webhook"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	port := flag.Int("port", 8000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	cfg := zap.NewProductionConfig()
	var hOpts *zapslog.HandlerOptions
	if *debug {
		// if debug is enabled, set the log level to debug and add source location
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		hOpts = &zapslog.HandlerOptions{
			AddSource: true,
		}
	}
	logger := zap.Must(cfg.Build())
	defer logger.Sync()
	slog.SetDefault(slog.New(zapslog.NewHandler(logger.Core(), hOpts)))
	slog.Info("kubeup", "version", version, "commit", commit, "date", date, "builtBy", builtBy)
	if *debug {
		slog.Warn("debug logging enabled, secrets will be written to stderr")
	}

	p, err := webhook.NewPublisher(publisherOptions()...)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	h, err := webhook.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		slog.Error("fatal error creating CloudEvent receiver", "error", err)
		os.Exit(1)
	}

	srvOpts := webhook.ServerOptions{
		Path: *path,
		Port: *port,
	}
	secEnv := getRequiredEnv("KU_SECRET_1", "KU_SECRET_2")
	switch {
	case secEnv != nil:
		srvOpts.Secret1 = secEnv["KU_SECRET_1"]
		srvOpts.Secret2 = secEnv["KU_SECRET_2"]
		slog.Info("protecting webhook with client secret")
	default:
		slog.Warn("no client secret configured, webhook will be unprotected")
	}

	s := webhook.NewServer(h, srvOpts)
	done := make(chan struct{})
	go webhook.Shutdown(context.Background(), s, done, 10*time.Second)

	slog.Info("starting server", "port", *port)
	if err = s.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server has unexpectedly shut down", "error", err)
		os.Exit(1)
	}
	slog.Info("waiting for server to shut down")
	<-done
	slog.Info("server has shut down")
}

func getRequiredEnv(vars ...string) map[string]string {
	envVars := make(map[string]string, 4)
	for _, k := range vars {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil
		}
		slog.Debug("using environment variable", "env", k, "value", v)
		envVars[k] = v
	}
	return envVars
}

func publisherOptions() []webhook.Options {
	opts := []webhook.Options{
		webhook.WithLogging(),
	}
	if mailEnv := getRequiredEnv("KU_EMAIL_FROM", "KU_EMAIL_TO", "KU_EMAIL_SUBJECT"); mailEnv != nil {
		opts = append(
			opts,
			webhook.WithEmail(mailEnv["KU_EMAIL_FROM"], mailEnv["KU_EMAIL_TO"], mailEnv["KU_EMAIL_SUBJECT"]))
	}
	if sgEnv := getRequiredEnv("KU_SENDGRID_APIKEY"); sgEnv != nil {
		opts = append(opts, webhook.WithSendgrid(sgEnv["KU_SENDGRID_APIKEY"]))
	}
	if smtpEnv := getRequiredEnv("KU_SMTP_HOST", "KU_SMTP_PORT", "KU_SMTP_USERNAME", "KU_SMTP_PASSWORD"); smtpEnv != nil {
		port, err := strconv.Atoi(smtpEnv["KU_SMTP_PORT"])
		if err != nil {
			// We just log the parsing error and continue.
			// webhook.WithSMTP will return an error if the port is 0.
			slog.Error("parsing SMTP port", "error", err)
		}
		opts = append(
			opts,
			webhook.WithSMTP(smtpEnv["KU_SMTP_HOST"], port, smtpEnv["KU_SMTP_USERNAME"], smtpEnv["KU_SMTP_PASSWORD"]))
	}
	return opts
}
