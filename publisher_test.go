package kubeup_test

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/joergjo/go-samples/kubeup"
)

func TestMailTemplate(t *testing.T) {
	ke := kubeup.NewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}
	vue := kubeup.VersionUpdateEvent{
		ResourceID:                         "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster",
		NewKubernetesVersionAvailableEvent: ke,
	}

	tmpl := template.Must(template.New("email").Parse(kubeup.TemplateEmail))
	var b bytes.Buffer
	err := tmpl.Execute(&b, &vue)
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
}
