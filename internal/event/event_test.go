package event_test

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"testing"

	"github.com/joergjo/kubeup/internal/event"
)

func TestDeserializeValidJSON(t *testing.T) {
	tests := []struct {
		name   string
		in     jsontext.Value
		target any
	}{
		{
			name:   "cluster-support-ended-event",
			in:     jsontext.Value(`{ "kubernetesVersion": "1.24.0" }`),
			target: &event.ContainerServiceClusterSupportEndedEvent{},
		},
		{
			name:   "cluster-support-ending-event",
			in:     jsontext.Value(`{ "kubernetesVersion": "1.24.0" }`),
			target: &event.ContainerServiceClusterSupportEndingEvent{},
		},
		{
			name: "new-kubernetes-version-available-event",
			in: jsontext.Value(`{
				"latestSupportedKubernetesVersion": "1.27.0",
				"latestStableKubernetesVersion": "1.26.0",
				"lowestMinorKubernetesVersion": "1.25.0",
				"latestPreviewKubernetesVersion": "1.28.0"
			}`),
			target: &event.ContainerServiceNewKubernetesVersionAvailableEvent{},
		},
		{
			name:   "node-pool-rolling-failed-event",
			in:     jsontext.Value(`{ "nodePoolName": "nodepool1" }`),
			target: &event.ContainerServiceNodePoolRollingFailedEvent{},
		},
		{
			name:   "node-pool-rolling-started-event",
			in:     jsontext.Value(`{ "nodePoolName": "nodepool1" }`),
			target: &event.ContainerServiceNodePoolRollingStartedEvent{},
		},
		{
			name:   "node-pool-rolling-succeeded-event",
			in:     jsontext.Value(`{ "nodePoolName": "nodepool1" }`),
			target: &event.ContainerServiceNodePoolRollingSucceededEvent{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := bytes.NewReader(tc.in)
			// The underlying event types in azsystemevents don't have exported JSON tags, so we need to match case-insensitive names.
			if got := json.UnmarshalRead(r, tc.target); got != nil {
				t.Errorf("error unmarshaling json %q: %v", tc.in, got)
			}
		})
	}
}

func TestDeserializeTrailingData(t *testing.T) {
	r := bytes.NewReader(jsontext.Value(`{ "nodePoolName": "nodepool1" } extra`))
	target := &event.ContainerServiceNodePoolRollingSucceededEvent{}
	if got := json.UnmarshalRead(r, target); got == nil {
		t.Errorf("expected error unmarshaling json with trailing data, got nil")
	}
}
