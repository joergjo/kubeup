package kubeup

import (
	"errors"
	"html/template"

	"github.com/rs/zerolog/log"
)

type options struct {
	sendgrid        *sendgridOptions
	smtp            *smtpOptions
	log             bool
	customPublisher PublisherFunc
}

type sendgridOptions struct {
	apiKey string
	EmailTemplate
}

type smtpOptions struct {
	host     string
	port     int
	username string
	password string
	EmailTemplate
}

type Options func(options *options) error

func WithLogging() Options {
	return func(options *options) error {
		options.log = true
		log.Debug().Msg("Configured log publisher")
		return nil
	}
}

func WithSendgrid(apiKey string, email EmailTemplate) Options {
	return func(options *options) error {
		if apiKey == "" {
			return errors.New("SendGrid API key required")
		}
		if email.From == "" {
			return errors.New("SendGrid from address required")
		}
		if email.To == "" {
			return errors.New("SendGrid to address required")
		}
		if email.Subject == "" {
			return errors.New("SendGrid subject required")
		}
		if email.Templ == nil {
			email.Templ = template.Must(template.New("email").Parse(TemplateEmail))
		}
		s := sendgridOptions{
			apiKey:        apiKey,
			EmailTemplate: email,
		}
		options.sendgrid = &s
		log.Debug().Msg("Configured SendGrid publisher")
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

func WithSMTP(host string, port int, username string, password string, email EmailTemplate) Options {
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
		if email.From == "" {
			return errors.New("SMTP from address required")
		}
		if email.To == "" {
			return errors.New("SMTP to address required")
		}
		if email.Subject == "" {
			return errors.New("SMTP subject required")
		}
		if email.Templ == nil {
			email.Templ = template.Must(template.New("email").Parse(TemplateEmail))
		}
		s := smtpOptions{
			host:          host,
			port:          port,
			username:      username,
			password:      password,
			EmailTemplate: email,
		}
		options.smtp = &s
		log.Debug().Msg("Configured SMTP publisher")
		return nil
	}
}
