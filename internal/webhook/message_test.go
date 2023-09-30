package webhook_test

import (
	"testing"

	"github.com/joergjo/go-samples/kubeup/internal/webhook"
)

func TestNewKubernetesVersionTemplate(t *testing.T) {
	e := webhook.ContainerServiceNewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "new-kubernetes-version.gohtml")
}

func TestSupportEndingTemplate(t *testing.T) {
	e := webhook.ContainerServiceClusterSupportEndingEvent{
		ContainerServiceClusterSupportEvent: webhook.ContainerServiceClusterSupportEvent{
			KubernetesVersion: "1.19.0",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "cluster-support-ending.gohtml")
}

func TestSupportEndedTemplate(t *testing.T) {
	e := webhook.ContainerServiceClusterSupportEndedEvent{
		ContainerServiceClusterSupportEvent: webhook.ContainerServiceClusterSupportEvent{
			KubernetesVersion: "1.19.0",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "cluster-support-ended.gohtml")
}

func TestNodePoolRollingStartedTemplate(t *testing.T) {
	e := webhook.ContainerServiceNodePoolRollingStartedEvent{
		ContainerServiceClusterRollingEvent: webhook.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-started.gohtml")
}

func TestNodePoolRollingSucceededTemplate(t *testing.T) {
	e := webhook.ContainerServiceNodePoolRollingSucceededEvent{
		ContainerServiceClusterRollingEvent: webhook.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-succeeded.gohtml")
}

func TestNodePoolRollingFailedTemplate(t *testing.T) {
	e := webhook.ContainerServiceNodePoolRollingFailedEvent{
		ContainerServiceClusterRollingEvent: webhook.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-failed.gohtml")
}

func testTemplate[T webhook.ContainerServiceEvent](t *testing.T, e T, src, filename string) {
	mb := webhook.NewMessageBuilder[T](filename)
	_, err := mb.Build(e, src)
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
}
