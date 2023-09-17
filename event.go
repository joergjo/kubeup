package kubeup

import (
	"fmt"
	"strings"
)

const (
	EventNewKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"
	EventClusterSupportEnding          = "Microsoft.ContainerService.ClusterSupportEnding"
	EventClusterSupportEnded           = "Microsoft.ContainerService.ClusterSupportEnded"
	EventNodePoolRollingStarted        = "Microsoft.ContainerService.NodePoolRollingStarted"
	EventNodePoolRollingSucceeded      = "Microsoft.ContainerService.NodePoolRollingSucceeded"
	EventNodePoolRollingFailed         = "Microsoft.ContainerService.NodePoolRollingFailed"
)

// ContainerServiceNewKubernetesVersionAvailableEvent is the event that is sent by Azure Kubernetes Service
// when a new Kubernetes version is available in the CloudEvent's data field.
type ContainerServiceNewKubernetesVersionAvailableEvent struct {
	LatestSupportedKubernetesVersion string `json:"latestSupportedKubernetesVersion"`
	LatestStableKubernetesVersion    string `json:"latestStableKubernetesVersion"`
	LowestMinorKubernetesVersion     string `json:"lowestMinorKubernetesVersion"`
	LatestPreviewKubernetesVersion   string `json:"latestPreviewKubernetesVersion"`
}

func (e ContainerServiceNewKubernetesVersionAvailableEvent) String() string {
	var b strings.Builder
	b.WriteString("New Kubernetes version available:\n")
	b.WriteString(fmt.Sprintf("Latest supported version: %s\n", e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest stable version: %s\n", e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("Lowest minor version: %s\n", e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest preview version: %s", e.LatestPreviewKubernetesVersion))
	return b.String()
}

type ContainerServiceClusterSupportEvent struct {
	KubernetesVersion string `json:"kubernetesVersion"`
}

func (e ContainerServiceClusterSupportEvent) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Kubernetes version: %s", e.KubernetesVersion))
	return b.String()
}

type ContainerServiceClusterSupportEndedEvent struct {
	ContainerServiceClusterSupportEvent
}

type ContainerServiceClusterSupportEndingEvent struct {
	ContainerServiceClusterSupportEvent
}

type ContainerServiceClusterRollingEvent struct {
	NodePoolName string `json:"nodePoolName"`
}

func (e ContainerServiceClusterRollingEvent) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Node pool name: %s", e.NodePoolName))
	return b.String()
}

type ContainerServiceNodePoolRollingStartedEvent struct {
	ContainerServiceClusterRollingEvent
}

type ContainerServiceNodePoolRollingSucceededEvent struct {
	ContainerServiceClusterRollingEvent
}

type ContainerServiceNodePoolRollingFailedEvent struct {
	ContainerServiceClusterRollingEvent
}

type ContainerServiceEvent interface {
	ContainerServiceNewKubernetesVersionAvailableEvent |
		ContainerServiceClusterSupportEndingEvent |
		ContainerServiceClusterSupportEndedEvent |
		ContainerServiceNodePoolRollingStartedEvent |
		ContainerServiceNodePoolRollingSucceededEvent |
		ContainerServiceNodePoolRollingFailedEvent
	fmt.Stringer
}
