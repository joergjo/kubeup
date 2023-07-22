package kubeup

import (
	"bytes"
	"html/template"

	"github.com/joergjo/go-samples/kubeup/templates"
)

type EmailTemplate struct {
	From    string
	To      string
	Subject string
	Tmpl    *template.Template
}

func NewEmailTemplate(from, to, subject, tmpl string) *EmailTemplate {
	t := EmailTemplate{
		From:    from,
		To:      to,
		Subject: subject,
		Tmpl:    template.Must(template.ParseFS(templates.FS, tmpl)),
	}
	return &t
}

func (e *EmailTemplate) Html(data any) (string, error) {
	var buf bytes.Buffer
	if err := e.Tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
