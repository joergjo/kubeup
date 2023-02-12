package kubeup

import (
	"fmt"
	"strings"
)

// NewKubernetesVersionAvailableEvent is the event that is sent by Azure Kubernetes Service
// when a new Kubernetes version is available in the CloudEvent's data field.
type NewKubernetesVersionAvailableEvent struct {
	LatestSupportedKubernetesVersion string `json:"latestSupportedKubernetesVersion"`
	LatestStableKubernetesVersion    string `json:"latestStableKubernetesVersion"`
	LowestMinorKubernetesVersion     string `json:"lowestMinorKubernetesVersion"`
	LatestPreviewKubernetesVersion   string `json:"latestPreviewKubernetesVersion"`
}

func (e NewKubernetesVersionAvailableEvent) String() string {
	var b strings.Builder
	b.WriteString("New Kubernetes version available:\n")
	b.WriteString(fmt.Sprintf("Latest supported version: %s\n", e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest stable version: %s\n", e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("Lowest minor version: %s\n", e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("Latest preview version: %s\n", e.LatestPreviewKubernetesVersion))
	return b.String()
}

// VersionUpdateEvent is the event that is sent to kubeup publishers. It embeds the
// NewKubernetesVersionAvailableEvent and adds the resource ID of the cluster.
type VersionUpdateEvent struct {
	ResourceID string
	NewKubernetesVersionAvailableEvent
}

func (e VersionUpdateEvent) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Resource ID: %s\n", e.ResourceID))
	b.WriteString(e.NewKubernetesVersionAvailableEvent.String())
	return b.String()
}
