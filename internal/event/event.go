package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func (e *ContainerServiceNewKubernetesVersionAvailableEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceNewKubernetesVersionAvailableEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.LatestSupportedKubernetesVersion == nil {
		return fmt.Errorf("missing required field: latestSupportedKubernetesVersion")
	}
	if in.LatestStableKubernetesVersion == nil {
		return fmt.Errorf("missing required field: latestStableKubernetesVersion")
	}
	if in.LowestMinorKubernetesVersion == nil {
		return fmt.Errorf("missing required field: lowestMinorKubernetesVersion")
	}
	if in.LatestPreviewKubernetesVersion == nil {
		return fmt.Errorf("missing required field: latestPreviewKubernetesVersion")
	}

	e.ContainerServiceNewKubernetesVersionAvailableEventData =
		azsystemevents.ContainerServiceNewKubernetesVersionAvailableEventData(in)
	return nil
}

// NewContainerServiceNewKubernetesVersionAvailableEvent creates a new event for Kubernetes version availability notifications.
// Parameters represent the different version numbers reported by Azure.
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

// NewContainerServiceClusterSupportEndingEvent creates a new event for Kubernetes version support ending notifications.
// The version parameter is the Kubernetes version for which support is ending.
func NewContainerServiceClusterSupportEndingEvent(version string) ContainerServiceClusterSupportEndingEvent {
	return ContainerServiceClusterSupportEndingEvent{
		ContainerServiceClusterSupportEndingEventData: azsystemevents.ContainerServiceClusterSupportEndingEventData{
			KubernetesVersion: &version,
		},
	}
}

func (e *ContainerServiceClusterSupportEndingEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceClusterSupportEndingEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.KubernetesVersion == nil {
		return fmt.Errorf("missing required field: kubernetesVersion")
	}

	e.ContainerServiceClusterSupportEndingEventData =
		azsystemevents.ContainerServiceClusterSupportEndingEventData(in)
	return nil
}

// ContainerServiceClusterSupportEndedEvent is the event sent when support for a Kubernetes version has ended.
type ContainerServiceClusterSupportEndedEvent struct {
	azsystemevents.ContainerServiceClusterSupportEndedEventData
}

// String returns a string representation of the ContainerServiceClusterSupportEndedEvent.
func (e ContainerServiceClusterSupportEndedEvent) String() string {
	return fmt.Sprintf("Support ended for Kubernetes version %s", *e.KubernetesVersion)
}

// NewContainerServiceClusterSupportEndedEvent creates a new event for Kubernetes version support ended notifications.
// The version parameter is the Kubernetes version for which support has ended.
func NewContainerServiceClusterSupportEndedEvent(version string) ContainerServiceClusterSupportEndedEvent {
	return ContainerServiceClusterSupportEndedEvent{
		ContainerServiceClusterSupportEndedEventData: azsystemevents.ContainerServiceClusterSupportEndedEventData{
			KubernetesVersion: &version,
		},
	}
}

func (e *ContainerServiceClusterSupportEndedEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceClusterSupportEndedEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.KubernetesVersion == nil {
		return fmt.Errorf("missing required field: kubernetesVersion")
	}

	e.ContainerServiceClusterSupportEndedEventData =
		azsystemevents.ContainerServiceClusterSupportEndedEventData(in)
	return nil
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

// NewContainerServiceNodePoolRollingStartedEvent creates a new event for node pool rolling upgrade started notifications.
// The name parameter is the name of the node pool being upgraded.
func NewContainerServiceNodePoolRollingStartedEvent(name string) ContainerServiceNodePoolRollingStartedEvent {
	return ContainerServiceNodePoolRollingStartedEvent{
		ContainerServiceNodePoolRollingStartedEventData: azsystemevents.ContainerServiceNodePoolRollingStartedEventData{
			NodePoolName: &name,
		},
	}
}

func (e *ContainerServiceNodePoolRollingStartedEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceNodePoolRollingStartedEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.NodePoolName == nil {
		return fmt.Errorf("missing required field: nodePoolName")
	}

	e.ContainerServiceNodePoolRollingStartedEventData =
		azsystemevents.ContainerServiceNodePoolRollingStartedEventData(in)
	return nil
}

// ContainerServiceNodePoolRollingSucceededEvent is the event sent when a node pool rolling upgrade has succeeded.
type ContainerServiceNodePoolRollingSucceededEvent struct {
	azsystemevents.ContainerServiceNodePoolRollingSucceededEventData
}

// String returns a string representation of the ContainerServiceNodePoolRollingSucceededEvent.
func (e ContainerServiceNodePoolRollingSucceededEvent) String() string {
	return fmt.Sprintf("Upgrade succeeded for node pool %s", *e.NodePoolName)
}

// NewContainerServiceNodePoolRollingSucceededEvent creates a new event for node pool rolling upgrade succeeded notifications.
// The name parameter is the name of the node pool that was successfully upgraded.
func NewContainerServiceNodePoolRollingSucceededEvent(name string) ContainerServiceNodePoolRollingSucceededEvent {
	return ContainerServiceNodePoolRollingSucceededEvent{
		ContainerServiceNodePoolRollingSucceededEventData: azsystemevents.ContainerServiceNodePoolRollingSucceededEventData{
			NodePoolName: &name,
		},
	}
}

func (e *ContainerServiceNodePoolRollingSucceededEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceNodePoolRollingSucceededEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.NodePoolName == nil {
		return fmt.Errorf("missing required field: nodePoolName")
	}

	e.ContainerServiceNodePoolRollingSucceededEventData =
		azsystemevents.ContainerServiceNodePoolRollingSucceededEventData(in)
	return nil
}

// ContainerServiceNodePoolRollingFailedEvent is the event sent when a node pool rolling upgrade has failed.
type ContainerServiceNodePoolRollingFailedEvent struct {
	azsystemevents.ContainerServiceNodePoolRollingFailedEventData
}

// String returns a string representation of the ContainerServiceNodePoolRollingFailedEvent.
func (e ContainerServiceNodePoolRollingFailedEvent) String() string {
	return fmt.Sprintf("Upgrade failed for node pool name %s", *e.NodePoolName)
}

// NewContainerServiceNodePoolRollingFailedEvent creates a new event for node pool rolling upgrade failed notifications.
// The name parameter is the name of the node pool whose upgrade failed.
func NewContainerServiceNodePoolRollingFailedEvent(name string) ContainerServiceNodePoolRollingFailedEvent {
	return ContainerServiceNodePoolRollingFailedEvent{
		ContainerServiceNodePoolRollingFailedEventData: azsystemevents.ContainerServiceNodePoolRollingFailedEventData{
			NodePoolName: &name,
		},
	}
}

func (e *ContainerServiceNodePoolRollingFailedEvent) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()

	type alias azsystemevents.ContainerServiceNodePoolRollingFailedEventData
	var in alias

	if err := dec.Decode(&in); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("invalid JSON: trailing data")
	}

	if in.NodePoolName == nil {
		return fmt.Errorf("missing required field: nodePoolName")
	}

	e.ContainerServiceNodePoolRollingFailedEventData =
		azsystemevents.ContainerServiceNodePoolRollingFailedEventData(in)
	return nil
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
