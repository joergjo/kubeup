package kubeup

import (
	"context"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	httpbinding "github.com/cloudevents/sdk-go/v2/protocol/http"
)

type newKubernetesVersionAvailableEvent struct {
	LatestSupportedKubernetesVersion string `json:"latestSupportedKubernetesVersion"`
	LatestStableKubernetesVersion    string `json:"latestStableKubernetesVersion"`
	LowestMinorKubernetesVersion     string `json:"lowestMinorKubernetesVersion"`
	LatestPreviewKubernetesVersion   string `json:"latestPreviewKubernetesVersion"`
}

const newKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"

func Run(ctx context.Context, path string, port int) error {
	c, err := cloudevents.NewClientHTTP(
		httpbinding.WithPath(path),
		httpbinding.WithPort(port),
		httpbinding.WithOptionsHandlerFunc(validate),
	)
	if err != nil {
		return err
	}

	log.Printf("Receiver using path %s, listening on port %d\n", path, port)
	return c.StartReceiver(ctx, receive)
}

func validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Webhook-Allowed-Origin", "eventgrid.azure.net")
	w.WriteHeader(http.StatusOK)
	log.Printf("Validated subscription request")
}

func receive(ctx context.Context, ev cloudevents.Event) protocol.Result {
	if ev.Type() != newKubernetesVersionAvailable {
		log.Printf("Received unexpected CloudEvent of type %q", ev.Type())
		return cloudevents.NewHTTPResult(http.StatusBadRequest, "unexpected CloudEvent type %q", ev.Type())
	}

	kev := newKubernetesVersionAvailableEvent{}
	if err := ev.DataAs(&kev); err != nil {
		log.Printf("Failed to deserialize NewKubernetesVersionAvailable data: %v\n", err)
		return cloudevents.NewHTTPResult(http.StatusBadRequest, "invalid NewKubernetesVersionAvailable data")
	}

	log.Printf("Latest supported: %s", kev.LatestSupportedKubernetesVersion)
	log.Printf("Latest stable: %s", kev.LatestStableKubernetesVersion)
	log.Printf("Lowest minor: %s", kev.LowestMinorKubernetesVersion)
	log.Printf("Latest preview: %s", kev.LatestPreviewKubernetesVersion)
	return cloudevents.NewHTTPResult(http.StatusOK, "")
}
