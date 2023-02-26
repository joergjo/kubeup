package kubeup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/joergjo/go-samples/kubeup"
)

func TestValidation(t *testing.T) {
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
			pub, _ := kubeup.NewPublisher()
			h, err := kubeup.NewCloudEventHandler(context.Background(), pub)
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
			res, err := c.Do(req)
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}

			if res.StatusCode != tc.status {
				t.Errorf("Want status code %d, got %d", tc.status, res.StatusCode)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	ke := kubeup.NewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
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
			name:        "valid_cloudevent",
			eventType:   kubeup.EventTypeNewKubernetesVersionAvailable,
			contentType: cloudevents.ApplicationCloudEventsJSON,
			method:      http.MethodPost,
			data:        ke,
			status:      http.StatusOK,
		},
		{
			name:        "invalid_event_type",
			eventType:   "invalid_event_type",
			contentType: cloudevents.ApplicationCloudEventsJSON,
			data:        ke,
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
			pub, _ := kubeup.NewPublisher()
			h, err := kubeup.NewCloudEventHandler(context.Background(), pub)
			if err != nil {
				t.Fatalf("Error creating handler: %v", err)
			}

			e := cloudevents.NewEvent()
			e.SetID("1234567890abcdef1234567890abcdef12345678")
			e.SetSource("/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster")
			e.SetType(tc.eventType)
			e.SetData(cloudevents.ApplicationCloudEventsJSON, ke)

			body, err := json.Marshal(e)
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
	ke := kubeup.NewKubernetesVersionAvailableEvent{
		LatestSupportedKubernetesVersion: "1.24.0",
		LatestStableKubernetesVersion:    "1.23.0",
		LowestMinorKubernetesVersion:     "1.22.0",
		LatestPreviewKubernetesVersion:   "1.25.0",
	}

	opts := kubeup.WithPublisherFunc(func(event kubeup.ResourceUpdateEvent) error {
		err1 := errors.New("first error publishing event")
		err2 := errors.New("second error publishing event")
		return errors.Join(err1, err2)
	})
	pub, _ := kubeup.NewPublisher(opts)
	h, err := kubeup.NewCloudEventHandler(context.Background(), pub)
	if err != nil {
		t.Fatalf("Error creating handler: %v", err)
	}
	e := cloudevents.NewEvent()
	e.SetID("1234567890abcdef1234567890abcdef12345678")
	e.SetSource("/subscriptions/a27b9009-b63f-4c18-b50b-b91985e03b69/resourceGroups/test/providers/Microsoft.ContainerService/managedClusters/test-cluster")
	e.SetType(kubeup.EventTypeNewKubernetesVersionAvailable)
	e.SetData(cloudevents.ApplicationCloudEventsJSON, ke)

	body, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Error marshalling event: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/cloudevents+json")

	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	want := http.StatusOK
	if res.Result().StatusCode != want {
		t.Errorf("Want status code %d, got %d", want, res.Result().StatusCode)
	}

}
