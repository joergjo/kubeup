package kubeup_test

import (
	"testing"

	"github.com/joergjo/go-samples/kubeup"
)

func TestNewKubernetesVersionTemplate(t *testing.T) {
	e := kubeup.ContainerServiceNewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "new-kubernetes-version.gohtml")
}

func TestSupportEndingTemplate(t *testing.T) {
	e := kubeup.ContainerServiceClusterSupportEndingEvent{
		ContainerServiceClusterSupportEvent: kubeup.ContainerServiceClusterSupportEvent{
			KubernetesVersion: "1.19.0",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "cluster-support-ending.gohtml")
}

func TestSupportEndedTemplate(t *testing.T) {
	e := kubeup.ContainerServiceClusterSupportEndedEvent{
		ContainerServiceClusterSupportEvent: kubeup.ContainerServiceClusterSupportEvent{
			KubernetesVersion: "1.19.0",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "cluster-support-ended.gohtml")
}

func TestNodePoolRollingStartedTemplate(t *testing.T) {
	e := kubeup.ContainerServiceNodePoolRollingStartedEvent{
		ContainerServiceClusterRollingEvent: kubeup.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-started.gohtml")
}

func TestNodePoolRollingSucceededTemplate(t *testing.T) {
	e := kubeup.ContainerServiceNodePoolRollingSucceededEvent{
		ContainerServiceClusterRollingEvent: kubeup.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-succeeded.gohtml")
}

func TestNodePoolRollingFailedTemplate(t *testing.T) {
	e := kubeup.ContainerServiceNodePoolRollingFailedEvent{
		ContainerServiceClusterRollingEvent: kubeup.ContainerServiceClusterRollingEvent{
			NodePoolName: "pool1",
		},
	}
	src := "/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster"
	testTemplate(t, e, src, "nodepool-rolling-failed.gohtml")
}

func testTemplate[T kubeup.ContainerServiceEvent](t *testing.T, e T, src, filename string) {
	mb := kubeup.NewMessageBuilder[T](filename)
	_, err := mb.Build(e, src)
	if err != nil {
		t.Fatalf("Expected nil err, got: %v", err)
	}
}
