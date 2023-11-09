package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/joergjo/go-samples/kubeup/internal/webhook"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Adjust zerlog's configuration so it mirrors the CloudEvents SDK log output
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.MessageFieldName = "msg"
	zerolog.TimestampFieldName = "ts"
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	port := flag.Int("port", 8000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
	}
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
			log.Fatal().Err(err).Msg("parsing SMTP port")
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
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	h, err := webhook.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error creating CloudEvent receiver")
	}

	s := webhook.New(h, *port, *path)
	done := make(chan struct{})
	go webhook.Shutdown(context.Background(), s, done, 10*time.Second)

	log.Info().Msgf("starting server on port %d", *port)
	err = s.ListenAndServe()
	log.Info().Msgf("waiting for server to shut down")
	<-done
	log.Info().Err(err).Msg("server has shut down")
}

func getEnvVars(vars ...string) map[string]string {
	envVars := make(map[string]string, 4)
	for _, k := range vars {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil
		}
		log.Debug().Msgf("%s=%q", k, v)
		envVars[k] = v
	}
	return envVars
}
