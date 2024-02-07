package webhook

import (
	"fmt"
	"strings"
)

const (
	// EventNewKubernetesVersionAvailable is the event type that is sent when a new Kubernetes version is available.
	EventNewKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"
	// EventClusterSupportEnding is the event type that is sent when support for a Kubernetes version is ending.
	EventClusterSupportEnding = "Microsoft.ContainerService.ClusterSupportEnding"
	// EventClusterSupportEnded is the event type that is sent when support for a Kubernetes version has ended.
	EventClusterSupportEnded = "Microsoft.ContainerService.ClusterSupportEnded"
	// EventNodePoolRollingStarted is the event type that is sent when a node pool rolling upgrade has started.
	EventNodePoolRollingStarted = "Microsoft.ContainerService.NodePoolRollingStarted"
	// EventNodePoolRollingSucceeded is the event type that is sent when a node pool rolling upgrade has succeeded.
	EventNodePoolRollingSucceeded = "Microsoft.ContainerService.NodePoolRollingSucceeded"
	// EventNodePoolRollingFailed is the event type that is sent when a node pool rolling upgrade has failed.
	EventNodePoolRollingFailed = "Microsoft.ContainerService.NodePoolRollingFailed"
)

// ContainerServiceNewKubernetesVersionAvailableEvent is the event that is sent by Azure Kubernetes Service
// when a new Kubernetes version is available in the CloudEvent's data field.
type ContainerServiceNewKubernetesVersionAvailableEvent struct {
	// LatestSupportedKubernetesVersion is the latest supported Kubernetes version.
	LatestSupportedKubernetesVersion string `json:"latestSupportedKubernetesVersion"`
	// LatestStableKubernetesVersion is the latest stable Kubernetes version.
	LatestStableKubernetesVersion string `json:"latestStableKubernetesVersion"`
	// LowestMinorKubernetesVersion is the lowest minor Kubernetes version.
	LowestMinorKubernetesVersion string `json:"lowestMinorKubernetesVersion"`
	// LatestPreviewKubernetesVersion is the latest preview Kubernetes version.
	LatestPreviewKubernetesVersion string `json:"latestPreviewKubernetesVersion"`
}

// String returns a string representation of the ContainerServiceNewKubernetesVersionAvailableEvent.
func (e ContainerServiceNewKubernetesVersionAvailableEvent) String() string {
	var b strings.Builder
	b.WriteString("New Kubernetes version available:\n")
	b.WriteString(fmt.Sprintf("Latest supported version: %s\n", e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest stable version: %s\n", e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("Lowest minor version: %s\n", e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest preview version: %s", e.LatestPreviewKubernetesVersion))
	return b.String()
}

// ContainerServiceClusterSupportEvent represents the commonality for support ending and support ended events.
type ContainerServiceClusterSupportEvent struct {
	KubernetesVersion string `json:"kubernetesVersion"`
}

// ContainerServiceClusterSupportEndingEvent is the event sent when support for a Kubernetes version is ending.
type ContainerServiceClusterSupportEndingEvent struct {
	ContainerServiceClusterSupportEvent
}

// String returns a string representation of the ContainerServiceClusterSupportEndingEvent.
func (e ContainerServiceClusterSupportEndingEvent) String() string {
	return fmt.Sprintf("Support ending for Kubernetes version %s", e.KubernetesVersion)
}

// ContainerServiceClusterSupportEndedEvent is the event sent when support for a Kubernetes version has ended.
type ContainerServiceClusterSupportEndedEvent struct {
	ContainerServiceClusterSupportEvent
}

// String returns a string representation of the ContainerServiceClusterSupportEndedEvent.
func (e ContainerServiceClusterSupportEndedEvent) String() string {
	return fmt.Sprintf("Support ended for Kubernetes version %s", e.KubernetesVersion)
}

// ContainerServiceClusterRollingEvent represents the commonality for node pool rolling events.
type ContainerServiceClusterRollingEvent struct {
	NodePoolName string `json:"nodePoolName"`
}

// ContainerServiceNodePoolRollingStartedEvent is the event sent when a node pool rolling upgrade has started.
type ContainerServiceNodePoolRollingStartedEvent struct {
	ContainerServiceClusterRollingEvent
}

// String returns a string representation of the ContainerServiceNodePoolRollingStartedEvent.
func (e ContainerServiceNodePoolRollingStartedEvent) String() string {
	return fmt.Sprintf("Upgrade started for node pool %s", e.NodePoolName)
}

// ContainerServiceNodePoolRollingSucceededEvent is the event sent when a node pool rolling upgrade has succeeded.
type ContainerServiceNodePoolRollingSucceededEvent struct {
	ContainerServiceClusterRollingEvent
}

// String returns a string representation of the ContainerServiceNodePoolRollingSucceededEvent.
func (e ContainerServiceNodePoolRollingSucceededEvent) String() string {
	return fmt.Sprintf("Upgrade succeeded for node pool %s", e.NodePoolName)
}

// ContainerServiceNodePoolRollingFailedEvent is the event sent when a node pool rolling upgrade has failed.
type ContainerServiceNodePoolRollingFailedEvent struct {
	ContainerServiceClusterRollingEvent
}

// String returns a string representation of the ContainerServiceNodePoolRollingFailedEvent.
func (e ContainerServiceNodePoolRollingFailedEvent) String() string {
	return fmt.Sprintf("Upgrade failed for node pool name %s", e.NodePoolName)
}

// ContainerServiceEvent is the constraint set of all possible events.
type ContainerServiceEvent interface {
	ContainerServiceNewKubernetesVersionAvailableEvent |
		ContainerServiceClusterSupportEndingEvent |
		ContainerServiceClusterSupportEndedEvent |
		ContainerServiceNodePoolRollingStartedEvent |
		ContainerServiceNodePoolRollingSucceededEvent |
		ContainerServiceNodePoolRollingFailedEvent
	fmt.Stringer
}
