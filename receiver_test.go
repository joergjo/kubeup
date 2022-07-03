package kubeup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/joergjo/go-samples/kubeup"
)

type stubNotifier struct{}

func (s stubNotifier) Notify(e kubeup.NewKubernetesVersionAvailableEvent) error {
	return nil
}

func TestReceiverValidation(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		status int
	}{
		{
			name:   "valid_origin",
			origin: kubeup.AzureEventGridOrigin,
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
			h, err := kubeup.NewCloudEventHandler(context.Background(), stubNotifier{})
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}

			mux := http.NewServeMux()
			mux.Handle("/webhook", h)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodOptions, ts.URL+"/webhook", nil)
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}

			if tc.origin != "" {
				req.Header.Set("WebHook-Request-Origin", tc.origin)
			}
			c := ts.Client()
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}

			if resp.StatusCode != tc.status {
				t.Errorf("Expected status code %d, got %d", tc.status, resp.StatusCode)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	testData := kubeup.NewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}

	tests := []struct {
		name        string
		eventType   string
		method      string
		data        interface{}
		contentType string
		status      int
	}{
		{
			name:        "valid_cloudevent",
			eventType:   kubeup.EventTypeNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        testData,
			status:      http.StatusOK,
		},
		{
			name:        "invalid_event_type",
			eventType:   "invalid_event_type",
			contentType: cloudevents.ApplicationCloudEventsJSON,
			data:        testData,
			method:      http.MethodPost,
			status:      http.StatusBadRequest,
		},
		{
			name:        "get_not_allowed",
			eventType:   kubeup.EventTypeNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodGet,
			status:      http.StatusMethodNotAllowed,
		},
		{
			name:        "delete_not_allowed",
			eventType:   kubeup.EventTypeNewKubernetesVersionAvailable,
			contentType: "",
			data:        nil,
			method:      http.MethodDelete,
			status:      http.StatusMethodNotAllowed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := kubeup.NewCloudEventHandler(context.Background(), stubNotifier{})
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}

			mux := http.NewServeMux()
			mux.Handle("/webhook", h)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			event := cloudevents.NewEvent()
			event.SetID("1234567890abcdef1234567890abcdef12345678")
			event.SetSource("/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster")
			event.SetType(tc.eventType)
			event.SetData(cloudevents.ApplicationCloudEventsJSON, testData)

			body, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("Error marshalling event: %v", err)
			}

			req, err := http.NewRequest(tc.method, ts.URL+"/webhook", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}
			req.Header.Set("Content-Type", tc.contentType)

			c := ts.Client()
			resp, err := c.Do(req)
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}

			if resp.StatusCode != tc.status {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			}
		})
	}
}
