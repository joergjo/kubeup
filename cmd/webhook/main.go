package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joergjo/kubeup/internal/event"
	"github.com/joergjo/kubeup/internal/webhook"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type config struct {
	Port           int    `env:"KU_PORT" envDefault:"8000"`
	Path           string `env:"KU_PATH" envDefault:"/webhook"`
	Debug          bool   `env:"KU_DEBUG" envDefault:"false"`
	EmailFrom      string `env:"KU_EMAIL_FROM"`
	EmailTo        string `env:"KU_EMAIL_TO"`
	EmailSubject   string `env:"KU_EMAIL_SUBJECT"`
	SMTPHost       string `env:"KU_SMTP_HOST"`
	SMTPPort       int    `env:"KU_SMTP_PORT" envDefault:"587"`
	SMTPUsername   string `env:"KU_SMTP_USERNAME"`
	SMTPPassword   string `env:"KU_SMTP_PASSWORD"`
	SendGridAPIKey string `env:"KU_SENDGRID_APIKEY"`
	TenantID       string `env:"KU_TENANT_ID"`
	AppID          string `env:"KU_APP_ID"`
	Secret1        string `env:"KU_SECRET_1"`
	Secret2        string `env:"KU_SECRET_2"`
}

func (c config) HasEntraID() bool {
	return c.TenantID != "" && c.AppID != ""
}

func (c config) HasClientSecrets() bool {
	return c.Secret1 != "" && c.Secret2 != ""
}

func (c config) HasEmail() bool {
	return c.EmailFrom != "" && c.EmailTo != "" && c.EmailSubject != ""
}

func (c config) HasSMTP() bool {
	return c.SMTPHost != "" && c.SMTPPort > 0 && c.SMTPUsername != "" && c.SMTPPassword != ""
}

func (c config) HasSendGrid() bool {
	return c.SendGridAPIKey != ""
}

var (
	version string
	commit  string
	date    string
	builtBy string
)

const (
	defaultPort = 8000
	defaultPath = "/webhook"
)

func main() {
	cfg := configure()
	os.Exit(run(cfg))
}

func configure() config {
	var port int
	var path string
	var debug bool

	flag.IntVar(&port, "port", defaultPort, "HTTP listen port")
	flag.StringVar(&path, "path", defaultPath, "WebHook path")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	var cfg config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("parsing environment variables", "error", err)
		os.Exit(1)
	}

	// Apply command line flags if they are not set in the environment
	if cfg.Port == defaultPort {
		cfg.Port = port
	}
	if cfg.Path == defaultPath {
		cfg.Path = path
	}
	if !cfg.Debug {
		cfg.Debug = debug
	}

	return cfg
}

func setDefaultLogger(debug bool) func() {
	cfg := zap.NewProductionConfig()
	opts := []zapslog.HandlerOption{}
	if debug {
		// if debug is enabled, set the log level to debug and add source location
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		opts = append(opts, zapslog.WithCaller(true))
	}
	logger := zap.Must(cfg.Build())
	slog.SetDefault(slog.New(zapslog.NewHandler(logger.Core(), opts...)))
	return func() {
		logger.Sync()
	}
}

func run(cfg config) int {
	flush := setDefaultLogger(cfg.Debug)
	defer flush()

	slog.Info("kubeup", "version", version, "commit", commit, "date", date, "builtBy", builtBy, "goVersion", runtime.Version(), "goMaxProcs", runtime.GOMAXPROCS(0))
	if cfg.Debug {
		slog.Warn("debug logging enabled")
	}

	p, err := event.NewPublisher(publisherOptions(cfg)...)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		return 1
	}

	h, err := webhook.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		slog.Error("creating CloudEvent receiver", "error", err)
		return 1
	}

	opts := append(serverOptions(cfg), webhook.WithPath(cfg.Path), webhook.WithPort(cfg.Port))
	s, err := webhook.NewServer(h, opts...)
	if err != nil {
		slog.Error("creating server", "error", err)
		return 1
	}

	errC := make(chan error, 1)
	go func() {
		slog.Info("starting webhook server", "port", cfg.Port, "path", cfg.Path)
		errC <- s.ListenAndServe()
	}()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	var exit int
	select {
	case err := <-errC:
		if err != http.ErrServerClosed {
			slog.Error("webhook server error", "error", err)
			exit = 1
		}
	case sig := <-sigC:
		signal.Stop(sigC)
		slog.Warn("received signal, shutting down", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		slog.Info("waiting for server to shut down")
		if err := s.Shutdown(ctx); err != nil {
			slog.Error("shutting down server", "error", err)
			if err := s.Close(); err != nil {
				slog.Error("forcefully closing server", "error", err)
			}
		}
		slog.Info("server has shut down")
	}
	return exit
}

func publisherOptions(cfg config) []event.Options {
	opts := []event.Options{
		event.WithLogging(),
	}
	if cfg.HasEmail() {
		opts = append(opts, event.WithEmail(cfg.EmailFrom, cfg.EmailTo, cfg.EmailSubject))
	}
	if cfg.HasSMTP() {
		opts = append(opts, event.WithSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword))
	}
	if cfg.HasSendGrid() {
		opts = append(opts, event.WithSendgrid(cfg.SendGridAPIKey))
	}
	return opts
}

func serverOptions(cfg config) []webhook.Options {
	var opts []webhook.Options
	// We prefer Entra ID over client secrets. Using both doesn't make sense, as client secrets
	// don't add value when using Entra ID.
	switch {
	case cfg.HasEntraID():
		opts = append(opts, webhook.WithEntraID(cfg.TenantID, cfg.AppID))
	case cfg.HasClientSecrets():
		opts = append(opts, webhook.WithClientSecret(cfg.Secret1, cfg.Secret2))
	}
	return opts
}
