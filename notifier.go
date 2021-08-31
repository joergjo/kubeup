package kubeup

import (
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Notifier interface {
	Notify(NewKubernetesVersionAvailableEvent) error
}

type LogNotifier struct{}

func (l LogNotifier) Notify(e NewKubernetesVersionAvailableEvent) error {
	log.Println(e)
	return nil
}

type SendGridNotifier struct {
	Client   *sendgrid.Client
	FromAddr string
	ToAddr   string
	Subject  string
}

func NewSendGridNotifier(apiKey, from, to, subject string) SendGridNotifier {
	return SendGridNotifier{
		Client:   sendgrid.NewSendClient(apiKey),
		FromAddr: from,
		ToAddr:   to,
		Subject:  subject,
	}
}

func (s SendGridNotifier) Notify(e NewKubernetesVersionAvailableEvent) error {
	from := mail.NewEmail("Kubeup", s.FromAddr)
	to := mail.NewEmail("Kubernetes administrator", s.ToAddr)
	sub := s.Subject
	txt := e.String()
	html := e.Html()
	msg := mail.NewSingleEmail(from, sub, to, txt, html)
	res, err := s.Client.Send(msg)
	if err != nil {
		log.Printf("Error sending E-mail: %v", err)
		return err
	}
	log.Printf("Result: %+v", res)
	log.Printf("Notified %s", s.ToAddr)
	return nil
}
