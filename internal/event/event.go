package event

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azsystemevents"
)

// ContainerServiceNewKubernetesVersionAvailableEvent is the event that is sent by Azure Kubernetes Service
// when a new Kubernetes version is available in the CloudEvent's data field.
type ContainerServiceNewKubernetesVersionAvailableEvent struct {
	azsystemevents.ContainerServiceNewKubernetesVersionAvailableEventData
}

// String returns a string representation of the ContainerServiceNewKubernetesVersionAvailableEvent.
func (e ContainerServiceNewKubernetesVersionAvailableEvent) String() string {
	var b strings.Builder
	b.WriteString("New Kubernetes version available:\n")
	b.WriteString(fmt.Sprintf("Latest supported version: %s\n", *e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest stable version: %s\n", *e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("Lowest minor version: %s\n", *e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest preview version: %s", *e.LatestPreviewKubernetesVersion))
	return b.String()
}

func NewContainerServiceNewKubernetesVersionAvailableEvent(latestSupported, latestStable, lowestMinor, latestPreview string) ContainerServiceNewKubernetesVersionAvailableEvent {
	return ContainerServiceNewKubernetesVersionAvailableEvent{
		ContainerServiceNewKubernetesVersionAvailableEventData: azsystemevents.ContainerServiceNewKubernetesVersionAvailableEventData{
			LatestSupportedKubernetesVersion: &latestSupported,
			LatestStableKubernetesVersion:    &latestStable,
			LowestMinorKubernetesVersion:     &lowestMinor,
			LatestPreviewKubernetesVersion:   &latestPreview,
		},
	}
}

// ContainerServiceClusterSupportEndingEvent is the event sent when support for a Kubernetes version is ending.
type ContainerServiceClusterSupportEndingEvent struct {
	azsystemevents.ContainerServiceClusterSupportEndingEventData
}

// String returns a string representation of the ContainerServiceClusterSupportEndingEvent.
func (e ContainerServiceClusterSupportEndingEvent) String() string {
	return fmt.Sprintf("Support ending for Kubernetes version %s", *e.KubernetesVersion)
}

func NewContainerServiceClusterSupportEndingEvent(version string) ContainerServiceClusterSupportEndingEvent {
	return ContainerServiceClusterSupportEndingEvent{
		ContainerServiceClusterSupportEndingEventData: azsystemevents.ContainerServiceClusterSupportEndingEventData{
			KubernetesVersion: &version,
		},
	}
}

// ContainerServiceClusterSupportEndedEvent is the event sent when support for a Kubernetes version has ended.
type ContainerServiceClusterSupportEndedEvent struct {
	azsystemevents.ContainerServiceClusterSupportEndedEventData
}

// String returns a string representation of the ContainerServiceClusterSupportEndedEvent.
func (e ContainerServiceClusterSupportEndedEvent) String() string {
	return fmt.Sprintf("Support ended for Kubernetes version %s", *e.KubernetesVersion)
}

func NewContainerServiceClusterSupportEndedEvent(version string) ContainerServiceClusterSupportEndedEvent {
	return ContainerServiceClusterSupportEndedEvent{
		ContainerServiceClusterSupportEndedEventData: azsystemevents.ContainerServiceClusterSupportEndedEventData{
			KubernetesVersion: &version,
		},
	}
}

// ContainerServiceClusterRollingEvent represents the commonality for node pool rolling events.
type ContainerServiceClusterRollingEvent struct {
	NodePoolName string `json:"nodePoolName"`
}

// ContainerServiceNodePoolRollingStartedEvent is the event sent when a node pool rolling upgrade has started.
type ContainerServiceNodePoolRollingStartedEvent struct {
	azsystemevents.ContainerServiceNodePoolRollingStartedEventData
}

// String returns a string representation of the ContainerServiceNodePoolRollingStartedEvent.
func (e ContainerServiceNodePoolRollingStartedEvent) String() string {
	return fmt.Sprintf("Upgrade started for node pool %s", *e.NodePoolName)
}

func NewContainerServiceNodePoolRollingStartedEvent(name string) ContainerServiceNodePoolRollingStartedEvent {
	return ContainerServiceNodePoolRollingStartedEvent{
		ContainerServiceNodePoolRollingStartedEventData: azsystemevents.ContainerServiceNodePoolRollingStartedEventData{
			NodePoolName: &name,
		},
	}
}

// ContainerServiceNodePoolRollingSucceededEvent is the event sent when a node pool rolling upgrade has succeeded.
type ContainerServiceNodePoolRollingSucceededEvent struct {
	azsystemevents.ContainerServiceNodePoolRollingSucceededEventData
}

// String returns a string representation of the ContainerServiceNodePoolRollingSucceededEvent.
func (e ContainerServiceNodePoolRollingSucceededEvent) String() string {
	return fmt.Sprintf("Upgrade succeeded for node pool %s", *e.NodePoolName)
}

func NewContainerServiceNodePoolRollingSucceededEvent(name string) ContainerServiceNodePoolRollingSucceededEvent {
	return ContainerServiceNodePoolRollingSucceededEvent{
		ContainerServiceNodePoolRollingSucceededEventData: azsystemevents.ContainerServiceNodePoolRollingSucceededEventData{
			NodePoolName: &name,
		},
	}
}

// ContainerServiceNodePoolRollingFailedEvent is the event sent when a node pool rolling upgrade has failed.
type ContainerServiceNodePoolRollingFailedEvent struct {
	azsystemevents.ContainerServiceNodePoolRollingFailedEventData
}

// String returns a string representation of the ContainerServiceNodePoolRollingFailedEvent.
func (e ContainerServiceNodePoolRollingFailedEvent) String() string {
	return fmt.Sprintf("Upgrade failed for node pool name %s", *e.NodePoolName)
}

func NewContainerServiceNodePoolRollingFailedEvent(name string) ContainerServiceNodePoolRollingFailedEvent {
	return ContainerServiceNodePoolRollingFailedEvent{
		ContainerServiceNodePoolRollingFailedEventData: azsystemevents.ContainerServiceNodePoolRollingFailedEventData{
			NodePoolName: &name,
		},
	}
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
