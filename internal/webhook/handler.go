package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azsystemevents"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/joergjo/kubeup/internal/event"
)

const (
	// AzureEventGridOrigin represents the origin string for Azure Event Grid.
	AzureEventGridOrigin = "eventgrid.azure.net"
)

// NewCloudEventHandler creates a new CloudEvent handler with the given Publisher.
func NewCloudEventHandler(ctx context.Context, pub *event.Publisher) (http.Handler, error) {
	oh := newOptionsHandler(
		[]string{http.MethodPost},
		cehttp.DefaultAllowedRate,
		[]string{AzureEventGridOrigin},
	)
	p, err := cloudevents.NewHTTP(cehttp.WithOptionsHandlerFunc(oh))
	if err != nil {
		slog.Error("creating protocol settings", "error", err)
		return nil, err
	}
	rh, err := cloudevents.NewHTTPReceiveHandler(ctx, p, newEventReceiver(pub))
	if err != nil {
		slog.Error("creating receive handler", "error", err)
		return nil, err
	}
	return rh, nil
}

// validateRequestOrigin checks if the provided origin is allowed based on the allowed origins list.
// Returns the matched origin and a boolean indicating if the origin is allowed.
func validateRequestOrigin(origin string, allowed []string) (string, bool) {
	slog.Info("validating origin", "origin", origin)
	for _, ao := range allowed {
		if ao == "*" {
			return ao, true
		}
		o := strings.TrimSpace(ao)
		if o == origin {
			return o, true
		}
	}
	return origin, false
}

// newOptionsHandler creates an HTTP handler for OPTIONS requests following the CloudEvents webhook spec.
// It validates the origin, sets allowed methods, and handles rate limiting headers.
func newOptionsHandler(methods []string, rate int, origins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		headers := make(http.Header)

		ro := r.Header.Get("WebHook-Request-Origin")
		if origin, ok := validateRequestOrigin(ro, origins); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			headers.Set("WebHook-Allowed-Origin", origin)
		}

		if _, ok := r.Header[http.CanonicalHeaderKey("WebHook-Request-Rate")]; ok {
			headers.Set("WebHook-Allowed-Rate", strconv.Itoa(rate))
		}

		if len(methods) > 0 {
			headers.Set("Allow", strings.Join(methods, ", "))
		} else {
			headers.Set("Allow", http.MethodPost)
		}

		for k := range headers {
			w.Header().Set(k, headers.Get(k))
		}
	}
}

// newEventReceiver creates a function that processes CloudEvents.
// It handles different event types from Azure Kubernetes Service and publishes them using the provided publisher.
func newEventReceiver(p *event.Publisher) func(context.Context, cloudevents.Event) protocol.Result {
	return func(ctx context.Context, e cloudevents.Event) protocol.Result {
		slog.Info("received event", "id", e.ID())
		switch e.Type() {
		case azsystemevents.TypeContainerServiceNewKubernetesVersionAvailable: //event.EventNewKubernetesVersionAvailable:
			return publishEvent[event.ContainerServiceNewKubernetesVersionAvailableEvent](e, p, "new-kubernetes-version.gohtml")
		case azsystemevents.TypeContainerServiceClusterSupportEnding: // event.EventClusterSupportEnding:
			return publishEvent[event.ContainerServiceClusterSupportEndingEvent](e, p, "cluster-support-ending.gohtml")
		case azsystemevents.TypeContainerServiceClusterSupportEnded: // event.EventClusterSupportEnded:
			return publishEvent[event.ContainerServiceClusterSupportEndedEvent](e, p, "cluster-support-ended.gohtml")
		case azsystemevents.TypeContainerServiceNodePoolRollingStarted: // event.EventNodePoolRollingStarted:
			return publishEvent[event.ContainerServiceNodePoolRollingStartedEvent](e, p, "nodepool-rolling-started.gohtml")
		case azsystemevents.TypeContainerServiceNodePoolRollingSucceeded: //event.EventNodePoolRollingSucceeded:
			return publishEvent[event.ContainerServiceNodePoolRollingSucceededEvent](e, p, "nodepool-rolling-succeeded.gohtml")
		case azsystemevents.TypeContainerServiceNodePoolRollingFailed: // event.EventNodePoolRollingFailed:
			return publishEvent[event.ContainerServiceNodePoolRollingFailedEvent](e, p, "nodepool-rolling-failed.gohtml")
		case azsystemevents.TypeSubscriptionDeleted: // event.EventSubscriptionDeleted:
			slog.Warn("event subscription deleted", "resource", e.Source())
			return cloudevents.NewHTTPResult(http.StatusOK, "")
		default:
			slog.Warn("unexpected CloudEvent type", "type", e.Type())
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unexpected CloudEvent type %q", e.Type())
		}
	}
}

// publishEvent processes a specific type of CloudEvent, creates a message using the appropriate template,
// and publishes it using the event publisher. Returns an HTTP result indicating success or failure.
func publishEvent[T event.ContainerServiceEvent](e cloudevents.Event, p *event.Publisher, filename string) protocol.Result {
	ce, err := unmarshal[T](e)
	if err != nil {
		slog.Error("deserializing event", "error", err, "type", e.Type())
		return cloudevents.NewHTTPResult(http.StatusBadRequest, "invalid %s data", e.Type())
	}
	mb := event.NewMessageBuilder[T](filename)
	msg, err := mb.Build(ce, e.Source())
	if err != nil {
		slog.Error("building message", "error", err)
		return cloudevents.NewHTTPResult(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	if err := p.Publish(msg); err != nil {
		slog.Error("publishing message", "error", err)
		return cloudevents.NewHTTPResult(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	return cloudevents.NewHTTPResult(http.StatusOK, "")
}

// unmarshal deserializes the CloudEvent data into the appropriate event type.
// Uses strict JSON decoding to ensure the data matches the expected structure.
func unmarshal[T event.ContainerServiceEvent](e cloudevents.Event) (T, error) {
	var data T
	dec := json.NewDecoder(bytes.NewReader(e.DataEncoded))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&data); err != nil {
		return data, err
	}
	return data, nil
}
