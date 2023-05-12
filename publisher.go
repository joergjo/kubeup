package kubeup

import (
	"errors"
	"fmt"

	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type PublisherFunc func(e ResourceUpdateEvent) error

type Publisher struct {
	publisherFns []PublisherFunc
}

func (p *Publisher) Publish(e ResourceUpdateEvent) error {
	var result error
	for _, pub := range p.publisherFns {
		if err := pub(e); err != nil {
			result = errors.Join(err)
		}
	}
	return result
}

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
		p.publisherFns = append(p.publisherFns, newSendGridPublisher(*options.sendgrid))
	}
	if options.smtp != nil {
		p.publisherFns = append(p.publisherFns, newSMTPPublisher(*options.smtp))
	}
	if options.customPublisher != nil {
		p.publisherFns = append(p.publisherFns, options.customPublisher)
	}

	return &p, nil
}

func newSendGridPublisher(s sendgridOptions) PublisherFunc {
	client := sendgrid.NewSendClient(s.apiKey)
	return func(e ResourceUpdateEvent) error {
		html, err := s.EmailTemplate.Html(e)
		if err != nil {
			return err
		}

		from := sgmail.NewEmail("Kubeup", s.From)
		to := sgmail.NewEmail("Kubernetes administrator", s.To)
		msg := sgmail.NewSingleEmail(from, s.Subject, to, e.String(), html)
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
	return func(e ResourceUpdateEvent) error {
		log.Info().Str("ResourceID", e.ResourceID).Msgf("%s", e.NewKubernetesVersionAvailableEvent)
		return nil
	}
}

func newSMTPPublisher(s smtpOptions) PublisherFunc {
	return func(e ResourceUpdateEvent) error {
		html, err := s.EmailTemplate.Html(e)
		if err != nil {
			return err
		}

		msg := mail.NewMessage()
		msg.SetHeader("From", s.From)
		msg.SetHeader("To", s.To)
		msg.SetHeader("Subject", s.Subject)
		msg.SetBody("text/plain", e.String())
		msg.AddAlternative("text/html", html)
		dialer := mail.NewDialer(s.host, s.port, s.username, s.password)
		err = dialer.DialAndSend(msg)
		if err != nil {
			return err
		}

		log.Debug().Str("Email", s.To).Msgf("SMTP notification successfully sent")
		return nil
	}
}
