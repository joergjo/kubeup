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

type ContainerServiceClusterSupportEndingEvent struct {
	ContainerServiceClusterSupportEvent
}

func (e ContainerServiceClusterSupportEndingEvent) String() string {
	return fmt.Sprintf("Support ending for Kubernetes version %s", e.KubernetesVersion)
}

type ContainerServiceClusterSupportEndedEvent struct {
	ContainerServiceClusterSupportEvent
}

func (e ContainerServiceClusterSupportEndedEvent) String() string {
	return fmt.Sprintf("Support ended for Kubernetes version %s", e.KubernetesVersion)
}

type ContainerServiceClusterRollingEvent struct {
	NodePoolName string `json:"nodePoolName"`
}

type ContainerServiceNodePoolRollingStartedEvent struct {
	ContainerServiceClusterRollingEvent
}

func (e ContainerServiceNodePoolRollingStartedEvent) String() string {
	return fmt.Sprintf("Upgrade started for node pool %s", e.NodePoolName)
}

type ContainerServiceNodePoolRollingSucceededEvent struct {
	ContainerServiceClusterRollingEvent
}

func (e ContainerServiceNodePoolRollingSucceededEvent) String() string {
	return fmt.Sprintf("Upgrade succeeded for node pool %s", e.NodePoolName)
}

type ContainerServiceNodePoolRollingFailedEvent struct {
	ContainerServiceClusterRollingEvent
}

func (e ContainerServiceNodePoolRollingFailedEvent) String() string {
	return fmt.Sprintf("Upgrade failed for node pool name %s", e.NodePoolName)
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
