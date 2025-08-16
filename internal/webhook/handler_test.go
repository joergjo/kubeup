package webhook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/joergjo/kubeup/internal/event"
	"github.com/joergjo/kubeup/internal/webhook"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		status int
	}{
		{
			name:   "valid_origin",
			origin: webhook.AzureEventGridOrigin,
			status: http.StatusOK,
		},
		{
			name:   "invalid_origin",
			origin: "invalid_origin",
			status: http.StatusBadRequest,
		},
		{
			name:   "missing_origin",
			origin: "",
			status: http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := event.NewPublisher()
			h, err := webhook.NewCloudEventHandler(context.Background(), p)
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}

			req := httptest.NewRequest(http.MethodOptions, "http://localhost:8000/webhook", nil)
			if tc.origin != "" {
				req.Header.Set("WebHook-Request-Origin", tc.origin)
			}
			res := httptest.NewRecorder()

			h.ServeHTTP(res, req)
			if res.Result().StatusCode != tc.status {
				t.Errorf("Want status code %d, got %d", tc.status, res.Result().StatusCode)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	newVersionEvent := event.ContainerServiceNewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}
	rollingEvent := event.ContainerServiceClusterRollingEvent{
		NodePoolName: "nodepool1",
	}
	supportEvent := event.ContainerServiceClusterSupportEvent{
		KubernetesVersion: "1.26.0",
	}

	tests := []struct {
		name        string
		eventType   string
		method      string
		data        any
		contentType string
		status      int
	}{
		{
			name:        "new_kubernetes_version_available",
			eventType:   event.EventNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        newVersionEvent,
			status:      http.StatusOK,
		},
		{
			name:        "nodepool_rolling_started",
			eventType:   event.EventNodePoolRollingStarted,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.ContainerServiceNodePoolRollingStartedEvent{
				ContainerServiceClusterRollingEvent: rollingEvent,
			},
			status: http.StatusOK,
		},
		{
			name:        "nodepool_rolling_succeeded",
			eventType:   event.EventNodePoolRollingSucceeded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.ContainerServiceNodePoolRollingSucceededEvent{
				ContainerServiceClusterRollingEvent: rollingEvent,
			},
			status: http.StatusOK,
		},
		{
			name:        "nodepool_rolling_failed",
			eventType:   event.EventNodePoolRollingFailed,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.ContainerServiceNodePoolRollingFailedEvent{
				ContainerServiceClusterRollingEvent: rollingEvent,
			},
			status: http.StatusOK,
		},
		{
			name:        "cluster_support_ending",
			eventType:   event.EventClusterSupportEnding,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.ContainerServiceClusterSupportEndingEvent{
				ContainerServiceClusterSupportEvent: supportEvent,
			},
			status: http.StatusOK,
		},
		{
			name:        "cluster_support_ended",
			eventType:   event.EventClusterSupportEnded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.ContainerServiceClusterSupportEndedEvent{
				ContainerServiceClusterSupportEvent: supportEvent,
			},
			status: http.StatusOK,
		},
		{
			name:        "invalid_event_type",
			eventType:   "invalid_event_type",
			contentType: cloudevents.ApplicationCloudEventsJSON,
			data:        newVersionEvent,
			method:      http.MethodPost,
			status:      http.StatusBadRequest,
		},
		{
			name:        "get_not_allowed",
			eventType:   event.EventNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodGet,
			status:      http.StatusMethodNotAllowed,
		},
		{
			name:        "delete_not_allowed",
			eventType:   event.EventNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodDelete,
			status:      http.StatusMethodNotAllowed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := event.NewPublisher()
			h, err := webhook.NewCloudEventHandler(context.Background(), p)
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}

			ce := cloudevents.NewEvent()
			ce.SetID("1234567890abcdef1234567890abcdef12345678")
			ce.SetSource("/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster")
			ce.SetType(tc.eventType)
			ce.SetData(cloudevents.ApplicationCloudEventsJSON, tc.data)

			body, err := json.Marshal(ce)
			if err != nil {
				t.Fatalf("Error marshalling event: %v", err)
			}

			req := httptest.NewRequest(tc.method, "/webhook", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tc.contentType)

			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			if res.Result().StatusCode != tc.status {
				t.Errorf("Want status code %d, got %d", tc.status, res.Result().StatusCode)
			}
		})
	}
}

func TestPublisherError(t *testing.T) {
	newVersionEvent := event.ContainerServiceNewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}

	opts := event.WithPublisherFunc(func(m event.Message) error {
		err1 := errors.New("first error publishing event")
		err2 := errors.New("second error publishing event")
		return errors.Join(err1, err2)
	})
	p, _ := event.NewPublisher(opts)
	h, err := webhook.NewCloudEventHandler(context.Background(), p)
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}
	ce := cloudevents.NewEvent()
	ce.SetID("1234567890abcdef1234567890abcdef12345678")
	ce.SetSource("/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster")
	ce.SetType(event.EventNewKubernetesVersionAvailable)
	ce.SetData(cloudevents.ApplicationCloudEventsJSON, newVersionEvent)

	body, err := json.Marshal(ce)
	if err != nil {
		t.Fatalf("Error marshalling event: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")

	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	want := http.StatusInternalServerError
	if res.Result().StatusCode != want {
		t.Errorf("Want status code %d, got %d", want, res.Result().StatusCode)
	}
}
