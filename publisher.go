package kubeup

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const defaultMailTemplate = `
<h1>New Kubernetes version available</h1>
<table>
<tr><td>Latest supported version</td><td>{{ .LatestSupportedKubernetesVersion }}</td></tr>
<tr><td>Latest stable version</td><td>{{ .LatestStableKubernetesVersion }}</td></tr>
<tr><td>Lowest minor version</td><td>{{ .LowestMinorKubernetesVersion }} </td></tr>
<tr><td>Latest preview version</td><td>{{ .LatestPreviewKubernetesVersion }}</td></tr>
</table>`

type publisher func(e NewKubernetesVersionAvailableEvent) error

type Publisher struct {
	publishers []publisher
}

func (p *Publisher) Publish(e NewKubernetesVersionAvailableEvent) error {
	var result error
	for _, pub := range p.publishers {
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
		publishers: []publisher{},
	}
	if options.log {
		p.publishers = append(p.publishers, newLogHandler())
	}
	if options.sendgrid != nil {
		p.publishers = append(p.publishers, newSendGridHandler(*options.sendgrid))
	}
	if options.customPublisher != nil {
		p.publishers = append(p.publishers, options.customPublisher)
	}

	return &p, nil
}

func newSendGridHandler(s sendgridOptions) publisher {
	client := sendgrid.NewSendClient(s.apiKey)
	return func(e NewKubernetesVersionAvailableEvent) error {
		b := make([]byte, 0, 512)
		buf := bytes.NewBuffer(b)
		if err := s.tmpl.Execute(buf, e); err != nil {
			return err
		}

		from := mail.NewEmail("Kubeup", s.from)
		to := mail.NewEmail("Kubernetes administrator", s.to)
		msg := mail.NewSingleEmail(from, s.sub, to, e.String(), buf.String())
		res, err := client.Send(msg)
		if err != nil {
			return err
		}
		if res.StatusCode < 200 && res.StatusCode >= 300 {
			err = fmt.Errorf("unexpected SendGrid HTTP status code %d, response %q", res.StatusCode, res.Body)
			return err
		}

		log.Debug().Msgf("SendGrid has notified %q", s.to)
		return nil
	}
}

func newLogHandler() publisher {
	return func(e NewKubernetesVersionAvailableEvent) error {
		log.Info().Msgf("%s", e)
		return nil
	}
}
