package event_test

import (
	"testing"

	"github.com/joergjo/kubeup/internal/event"
)

func TestTemplates(t *testing.T) {
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"

	t.Run("new_kubernetes_version", func(t *testing.T) {
		testTypedTemplate(
			t,
			"new-kubernetes-version.gohtml",
			event.NewContainerServiceNewKubernetesVersionAvailableEvent(
				"1.24.0", "1.23.0", "1.22.0", "1.25.0"),
			src,
		)
	})

	t.Run("support_ending", func(t *testing.T) {
		testTypedTemplate(
			t,
			"cluster-support-ending.gohtml",
			event.NewContainerServiceClusterSupportEndingEvent("1.19.0"),
			src,
		)
	})

	t.Run("support_ended", func(t *testing.T) {
		testTypedTemplate(
			t,
			"cluster-support-ended.gohtml",
			event.NewContainerServiceClusterSupportEndedEvent("1.19.0"),
			src,
		)
	})

	t.Run("nodepool_rolling_started", func(t *testing.T) {
		testTypedTemplate(
			t,
			"nodepool-rolling-started.gohtml",
			event.NewContainerServiceNodePoolRollingStartedEvent("nodepool1"),
			src,
		)
	})

	t.Run("nodepool_rolling_succeeded", func(t *testing.T) {
		testTypedTemplate(
			t,
			"nodepool-rolling-succeeded.gohtml",
			event.NewContainerServiceNodePoolRollingSucceededEvent("nodepool1"),
			src,
		)
	})

	t.Run("nodepool_rolling_failed", func(t *testing.T) {
		testTypedTemplate(
			t,
			"nodepool-rolling-failed.gohtml",
			event.NewContainerServiceNodePoolRollingFailedEvent("nodepool1"),
			src,
		)
	})
}

// testTypedTemplate is a generic helper function to test each template
func testTypedTemplate[T event.ContainerServiceEvent](t *testing.T, templ string, e T, src string) {
	mb := event.NewMessageBuilder[T](templ)
	_, err := mb.Build(e, src)
	if err != nil {
		t.Fatalf("Expected nil err for template %s, got: %v", templ, err)
	}
}
