package webhook

import (
	"errors"

	"github.com/rs/zerolog/log"
)

type options struct {
	sendgrid        *sendgridOptions
	smtp            *smtpOptions
	email           *emailOptions
	log             bool
	customPublisher PublisherFunc
}

type sendgridOptions struct {
	apiKey string
}

type smtpOptions struct {
	host     string
	port     int
	username string
	password string
}

type emailOptions struct {
	from    string
	to      string
	subject string
}

type Options func(options *options) error

func WithLogging() Options {
	return func(options *options) error {
		options.log = true
		log.Debug().Msg("Configured log publisher")
		return nil
	}
}

func WithEmail(from, to, subject string) Options {
	return func(options *options) error {
		if from == "" {
			return errors.New("email from address required")
		}
		if to == "" {
			return errors.New("email to address required")
		}
		if subject == "" {
			return errors.New("email subject required")
		}
		e := emailOptions{from: from, to: to, subject: subject}
		options.email = &e
		log.Debug().Msg("Configured email")
		return nil
	}
}

func WithPublisherFunc(fn PublisherFunc) Options {
	return func(options *options) error {
		if fn == nil {
			return errors.New("PublisherFunc must not be nil")
		}
		options.customPublisher = fn
		log.Debug().Msg("Configured custom publisher")
		return nil
	}
}

func WithSendgrid(apiKey string) Options {
	return func(options *options) error {
		if apiKey == "" {
			return errors.New("SendGrid API key required")
		}
		s := sendgridOptions{
			apiKey: apiKey,
		}
		options.sendgrid = &s
		log.Debug().Msg("Configured SendGrid publisher")
		return nil
	}
}

func WithSMTP(host string, port int, username string, password string) Options {
	return func(options *options) error {
		if host == "" {
			return errors.New("SMTP host required")
		}
		if port == 0 {
			return errors.New("SMTP port required")
		}
		if username == "" {
			return errors.New("SMTP username required")
		}
		if password == "" {
			return errors.New("SMTP password required")
		}
		s := smtpOptions{
			host:     host,
			port:     port,
			username: username,
			password: password,
		}
		options.smtp = &s
		log.Debug().Msg("Configured SMTP publisher")
		return nil
	}
}
