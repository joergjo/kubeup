package webhook

import (
	"errors"
	"fmt"

	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// PublisherFunc is a function that publishes a message.
type PublisherFunc func(m Message) error

// Publisher is a message publisher.
type Publisher struct {
	publisherFns []PublisherFunc
}

// Publish sends a message to all registered publishers.
func (p *Publisher) Publish(m Message) error {
	var result error
	for _, pub := range p.publisherFns {
		if err := pub(m); err != nil {
			result = errors.Join(err)
		}
	}
	return result
}

// NewPublisher creates a new Publisher with the given options.
func NewPublisher(opts ...Options) (*Publisher, error) {
	var options options
	var err error

	for _, opt := range opts {
		if err = opt(&options); err != nil {
			return nil, fmt.Errorf("invalid options: %w", err)
		}
	}
	p := Publisher{
		publisherFns: []PublisherFunc{},
	}
	if options.log {
		p.publisherFns = append(p.publisherFns, newLogPublisher())
	}
	if options.sendgrid != nil {
		if options.email == nil {
			return nil, errors.New("email options required with SendGrid")
		}
		p.publisherFns = append(p.publisherFns, newSendGridPublisher(*options.sendgrid, *options.email))
	}
	if options.smtp != nil {
		if options.email == nil {
			return nil, errors.New("email options required with SMTP")
		}
		p.publisherFns = append(p.publisherFns, newSMTPPublisher(*options.smtp, *options.email))
	}
	if options.customPublisher != nil {
		p.publisherFns = append(p.publisherFns, options.customPublisher)
	}
	return &p, nil
}

func newSendGridPublisher(s sendgridOptions, e emailOptions) PublisherFunc {
	client := sendgrid.NewSendClient(s.apiKey)
	return func(m Message) error {
		from := sgmail.NewEmail("Kubeup", e.from)
		to := sgmail.NewEmail("Kubernetes administrator", e.to)
		msg := sgmail.NewSingleEmail(from, e.subject, to, m.PlainText, m.HTML)
		res, err := client.Send(msg)
		if err != nil {
			return err
		}
		if res.StatusCode < 200 && res.StatusCode >= 300 {
			return fmt.Errorf("unexpected SendGrid HTTP status code %d, response %q", res.StatusCode, res.Body)
		}
		log.Debug().Str("Email", to.Address).Msgf("SendGrid notification successfully sent")
		return nil
	}
}

func newLogPublisher() PublisherFunc {
	return func(m Message) error {
		log.Info().Str("Source", m.Source).Msg(m.PlainText)
		return nil
	}
}

func newSMTPPublisher(s smtpOptions, e emailOptions) PublisherFunc {
	return func(m Message) error {
		msg := mail.NewMessage()
		msg.SetHeader("From", e.from)
		msg.SetHeader("To", e.to)
		msg.SetHeader("Subject", e.subject)
		msg.SetBody("text/plain", m.PlainText)
		msg.AddAlternative("text/html", m.HTML)
		dialer := mail.NewDialer(s.host, s.port, s.username, s.password)
		if err := dialer.DialAndSend(msg); err != nil {
			return err
		}
		log.Debug().Str("Email", e.to).Msgf("SMTP notification successfully sent")
		return nil
	}
}
