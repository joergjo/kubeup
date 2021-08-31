package kubeup

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	httpbinding "github.com/cloudevents/sdk-go/v2/protocol/http"
)

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

func (e NewKubernetesVersionAvailableEvent) Html() string {
	var b strings.Builder
	b.WriteString("<h1>New Kubernetes version available</h1>")
	b.WriteString("<table>")
	b.WriteString(fmt.Sprintf("<tr><td>Latest supported version</td><td>%s</td></tr>", e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Latest stable version</td><td>%s</td></tr>", e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Lowest minor version</td><td>%s</td></tr>", e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Latest preview version</td><td>%s</td></tr>", e.LatestPreviewKubernetesVersion))
	b.WriteString("</table")
	return b.String()
}

const newKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"

func Run(ctx context.Context, path string, port int, notify Notifier) error {
	c, err := cloudevents.NewClientHTTP(
		httpbinding.WithPath(path),
		httpbinding.WithPort(port),
		httpbinding.WithOptionsHandlerFunc(validate),
	)
	if err != nil {
		log.Printf("Error creating receiver: %v", err)
		return err
	}

	log.Printf("Receiver using path %s, listening on port %d", path, port)
	return c.StartReceiver(ctx, newReceiveHandler(notify))
}

func validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Webhook-Allowed-Origin", "eventgrid.azure.net")
	w.WriteHeader(http.StatusOK)
	log.Printf("Validated subscription request")
}

func newReceiveHandler(n Notifier) func(context.Context, cloudevents.Event) protocol.Result {
	return func(ctx context.Context, e cloudevents.Event) protocol.Result {
		if e.Type() != newKubernetesVersionAvailable {
			log.Printf("Received unexpected CloudEvent of type %q", e.Type())
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unexpected CloudEvent type %q", e.Type())
		}

		ke := NewKubernetesVersionAvailableEvent{}
		if err := e.DataAs(&ke); err != nil {
			log.Printf("Failed to deserialize NewKubernetesVersionAvailable data: %v", err)
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "invalid NewKubernetesVersionAvailable data")
		}

		log.Printf("Received event with id %q", e.ID())
		if err := n.Notify(ke); err != nil {
			log.Printf("Failed to notify, event will be dropped. Error: %v", err)
		}
		return cloudevents.NewHTTPResult(http.StatusOK, "")
	}
}
