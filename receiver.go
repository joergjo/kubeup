package kubeup

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

const (
	EventTypeNewKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"
	AzureEventGridOrigin                   = "eventgrid.azure.net"
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

// TODO: Use templates instead of hardcoded strings
func (e NewKubernetesVersionAvailableEvent) Html() string {
	var b strings.Builder
	b.WriteString("<h1>New Kubernetes version available</h1>")
	b.WriteString("<table>")
	b.WriteString(fmt.Sprintf("<tr><td>Latest supported version</td><td>%s</td></tr>", e.LatestSupportedKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Latest stable version</td><td>%s</td></tr>", e.LatestStableKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Lowest minor version</td><td>%s</td></tr>", e.LowestMinorKubernetesVersion))
	b.WriteString(fmt.Sprintf("<tr><td>Latest preview version</td><td>%s</td></tr>", e.LatestPreviewKubernetesVersion))
	b.WriteString("</table>")
	return b.String()
}

func NewCloudEventHandler(ctx context.Context, n Notifier) (http.Handler, error) {
	p, err := cloudevents.NewHTTP(cehttp.WithDefaultOptionsHandlerFunc(
		[]string{http.MethodOptions},
		cehttp.DefaultAllowedRate,
		[]string{"eventgrid.azure.net"},
		false))
	if err != nil {
		log.Printf("Error creating protocol settings: %v", err)
		return nil, err
	}
	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, newReceiveHandler(n))
	if err != nil {
		log.Printf("Error creating receiver: %v", err)
		return nil, err
	}
	return h, nil
}

func newReceiveHandler(n Notifier) func(context.Context, cloudevents.Event) protocol.Result {
	return func(ctx context.Context, e cloudevents.Event) protocol.Result {
		if e.Type() != EventTypeNewKubernetesVersionAvailable {
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
