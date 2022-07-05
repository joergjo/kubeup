package kubeup

import (
	"bytes"
	"html/template"

	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Notifier interface {
	Notify(NewKubernetesVersionAvailableEvent) error
}

type LogNotifier struct{}

func (l LogNotifier) Notify(e NewKubernetesVersionAvailableEvent) error {
	log.Info().Msgf("%s", e)
	return nil
}

type SendGridNotifier struct {
	Client   *sendgrid.Client
	From     string
	To       string
	Subject  string
	Template *template.Template
}

const defaultMailTemplate = `
<h1>New Kubernetes version available</h1>
<table>
<tr><td>Latest supported version</td><td>{{ .LatestSupportedKubernetesVersion }}</td></tr>
<tr><td>Latest stable version</td><td>{{ .LatestStableKubernetesVersion }}</td></tr>
<tr><td>Lowest minor version</td><td>{{ .LowestMinorKubernetesVersion }} </td></tr>
<tr><td>Latest preview version</td><td>{{ .LatestPreviewKubernetesVersion }}</td></tr>
</table>`

func NewSendGridNotifier(apiKey, from, to, subject string, tmpl *template.Template) SendGridNotifier {
	if tmpl == nil {
		tmpl = template.Must(template.New("email").Parse(defaultMailTemplate))
	}
	return SendGridNotifier{
		Client:   sendgrid.NewSendClient(apiKey),
		From:     from,
		To:       to,
		Subject:  subject,
		Template: tmpl,
	}
}

func (s SendGridNotifier) Notify(e NewKubernetesVersionAvailableEvent) error {
	b := make([]byte, 0, 512)
	buf := bytes.NewBuffer(b)
	if err := s.Template.Execute(buf, e); err != nil {
		log.Error().Err(err).Msg("Error rendering HTML template")
		return err
	}

	from := mail.NewEmail("Kubeup", s.From)
	to := mail.NewEmail("Kubernetes administrator", s.To)
	msg := mail.NewSingleEmail(from, s.Subject, to, e.String(), buf.String())
	res, err := s.Client.Send(msg)
	if err != nil {
		log.Error().Err(err).Msg("Error sending E-mail")
		return err
	}
	log.Printf("Succesfully notified %q, SendGrid status code %d", s.To, res.StatusCode)
	return nil
}
