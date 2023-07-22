package kubeup_test

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/joergjo/go-samples/kubeup"
	"github.com/joergjo/go-samples/kubeup/templates"
)

func TestMailTemplate(t *testing.T) {
	ke := kubeup.NewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}
	vue := kubeup.ResourceUpdateEvent{
		ResourceID:                         "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster",
		NewKubernetesVersionAvailableEvent: ke,
	}

	tmpl := template.Must(template.ParseFS(templates.FS, "resourceUpdate.gohtml"))
	var b bytes.Buffer
	err := tmpl.Execute(&b, &vue)
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
}
