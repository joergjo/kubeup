package webhook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azsystemevents"
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
			eventType:   azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: event.NewContainerServiceNewKubernetesVersionAvailableEvent(
				"1.24.0",
				"1.23.0",
				"1.22.0",
				"1.25.0"),
			status: http.StatusOK,
		},
		{
			name:        "nodepool_rolling_started",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingStarted,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        event.NewContainerServiceNodePoolRollingStartedEvent("nodepool1"),
			status:      http.StatusOK,
		},
		{
			name:        "nodepool_rolling_succeeded",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingSucceeded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        event.NewContainerServiceNodePoolRollingSucceededEvent("nodepool1"),
			status:      http.StatusOK,
		},
		{
			name:        "nodepool_rolling_failed",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingFailed,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        event.NewContainerServiceNodePoolRollingFailedEvent("nodepool1"),
			status:      http.StatusOK,
		},
		{
			name:        "cluster_support_ending",
			eventType:   azsystemevents.TypeContainerServiceClusterSupportEnding,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        event.NewContainerServiceClusterSupportEndingEvent("1.19.0"),
			status:      http.StatusOK,
		},
		{
			name:        "cluster_support_ended",
			eventType:   azsystemevents.TypeContainerServiceClusterSupportEnded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        event.NewContainerServiceClusterSupportEndedEvent("1.19.0"),
			status:      http.StatusOK,
		},
		{
			name:        "invalid_event_type",
			eventType:   "invalid_event_type",
			contentType: cloudevents.ApplicationCloudEventsJSON,
			data: event.NewContainerServiceNewKubernetesVersionAvailableEvent(
				"1.24.0",
				"1.23.0",
				"1.22.0",
				"1.25.0"),
			method: http.MethodPost,
			status: http.StatusBadRequest,
		},
		{
			name:        "get_not_allowed",
			eventType:   azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodGet,
			status:      http.StatusMethodNotAllowed,
		},
		{
			name:        "delete_not_allowed",
			eventType:   azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodDelete,
			status:      http.StatusMethodNotAllowed,
		},
		{
			name:        "new_kubernetes_version_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: map[string]any{
				// missing latestSupportedKubernetesVersion
				"latestStableKubernetesVersion":  "1.23.0",
				"lowestMinorKubernetesVersion":   "1.22.0",
				"latestPreviewKubernetesVersion": "1.25.0",
			},
			status: http.StatusBadRequest,
		},
		{
			name:        "new_kubernetes_version_unknown_field",
			eventType:   azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data: map[string]any{
				"latestSupportedKubernetesVersion": "1.24.0",
				"latestStableKubernetesVersion":    "1.23.0",
				"lowestMinorKubernetesVersion":     "1.22.0",
				"latestPreviewKubernetesVersion":   "1.25.0",
				"someUnexpectedField":              "boom",
			},
			status: http.StatusBadRequest,
		},
		{
			name:        "cluster_support_ending_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceClusterSupportEnding,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        map[string]any{},
			status:      http.StatusBadRequest,
		},
		{
			name:        "cluster_support_ended_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceClusterSupportEnded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        map[string]any{},
			status:      http.StatusBadRequest,
		},
		{
			name:        "nodepool_rolling_started_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingStarted,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        map[string]any{},
			status:      http.StatusBadRequest,
		},
		{
			name:        "nodepool_rolling_succeeded_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingSucceeded,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        map[string]any{},
			status:      http.StatusBadRequest,
		},
		{
			name:        "nodepool_rolling_failed_missing_required_field",
			eventType:   azsystemevents.TypeContainerServiceNodePoolRollingFailed,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        map[string]any{},
			status:      http.StatusBadRequest,
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
	e := event.NewContainerServiceNewKubernetesVersionAvailableEvent(
		"1.24.0",
		"1.23.0",
		"1.22.0",
		"1.25.0")

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
	ce.SetType(azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable)
	ce.SetData(cloudevents.ApplicationCloudEventsJSON, e)

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
