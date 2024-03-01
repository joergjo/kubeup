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

	var opts []webhook.Options = []webhook.Options{
		webhook.WithLogging(),
	}
	if envVars := getEnvVars("KU_EMAIL_FROM",
		"KU_EMAIL_TO",
		"KU_EMAIL_SUBJECT"); envVars != nil {
		opts = append(
			opts,
			webhook.WithEmail(
				envVars["KU_EMAIL_FROM"],
				envVars["KU_EMAIL_TO"],
				envVars["KU_EMAIL_SUBJECT"]))
	}
	if envVars := getEnvVars("KU_SENDGRID_APIKEY"); envVars != nil {
		opts = append(
			opts,
			webhook.WithSendgrid(envVars["KU_SENDGRID_APIKEY"]))
	}
	if envVars := getEnvVars("KU_SMTP_HOST",
		"KU_SMTP_PORT",
		"KU_SMTP_USERNAME",
		"KU_SMTP_PASSWORD"); envVars != nil {
		port, err := strconv.Atoi(envVars["KU_SMTP_PORT"])
		if err != nil {
			slog.Error("parsing SMTP port", "error", err)
		}
		opts = append(
			opts,
			webhook.WithSMTP(
				envVars["KU_SMTP_HOST"],
				port,
				envVars["KU_SMTP_USERNAME"],
				envVars["KU_SMTP_PASSWORD"]))
	}

	p, err := webhook.NewPublisher(opts...)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	h, err := webhook.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		slog.Error("fatal error creating CloudEvent receiver", "error", err)
		os.Exit(1)
	}

	s := webhook.NewServer(h, *port, *path)
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

func getEnvVars(vars ...string) map[string]string {
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
