package kubeup

import (
	"errors"
	"html/template"

	"github.com/rs/zerolog/log"
)

type options struct {
	sendgrid        *sendgridOptions
	log             bool
	customPublisher publisher
}

type sendgridOptions struct {
	apiKey string
	from   string
	to     string
	sub    string
	tmpl   *template.Template
}

type Options func(options *options) error

func WithLogging() Options {
	return func(options *options) error {
		options.log = true
		log.Debug().Msg("Configured log publisher")
		return nil
	}
}

func WithSendgrid(apiKey, from, to, sub string, tmpl *template.Template) Options {
	return func(options *options) error {
		if apiKey == "" {
			return errors.New("API key required")
		}
		if from == "" {
			return errors.New("from address required")
		}
		if to == "" {
			return errors.New("to address required")
		}
		if sub == "" {
			return errors.New("subject required")
		}
		if tmpl == nil {
			tmpl = template.Must(template.New("email").Parse(TemplateEmail))
		}
		s := sendgridOptions{
			apiKey: apiKey,
			from:   from,
			to:     to,
			sub:    sub,
			tmpl:   tmpl,
		}
		options.sendgrid = &s
		log.Debug().Msg("Configured SendGrid publisher")
		return nil
	}
}

func WithPublisherFunc(fn func(e VersionUpdateEvent) error) Options {
	return func(options *options) error {
		if fn == nil {
			return errors.New("publisher func must not be nil")
		}
		options.customPublisher = fn
		log.Debug().Msg("Configured custom publisher")
		return nil
	}
}
