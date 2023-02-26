package kubeup

import (
	"bytes"
	"html/template"
)

type EmailTemplate struct {
	From    string
	To      string
	Subject string
	Templ   *template.Template
}

func (e EmailTemplate) Html(ve ResourceUpdateEvent) (string, error) {
	var buf bytes.Buffer
	if err := e.Templ.Execute(&buf, ve); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const TemplateEmail = `
<h1>New Kubernetes version available</h1>
<h2>Resource ID: {{ .ResourceID }}</h2>
<table>
<tr><td>Latest supported version</td><td>{{ .LatestSupportedKubernetesVersion }}</td></tr>
<tr><td>Latest stable version</td><td>{{ .LatestStableKubernetesVersion }}</td></tr>
<tr><td>Lowest minor version</td><td>{{ .LowestMinorKubernetesVersion }} </td></tr>
<tr><td>Latest preview version</td><td>{{ .LatestPreviewKubernetesVersion }}</td></tr>
</table>`
