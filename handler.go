package kubeup

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/rs/zerolog/log"
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

func NewCloudEventHandler(ctx context.Context, pub *Publisher) (http.Handler, error) {
	p, err := cloudevents.NewHTTP(cehttp.WithDefaultOptionsHandlerFunc(
		[]string{http.MethodOptions},
		cehttp.DefaultAllowedRate,
		[]string{"eventgrid.azure.net"},
		true))
	if err != nil {
		log.Error().Err(err).Msg("Error creating protocol settings")
		return nil, err
	}
	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, newEventReceiver(pub))
	if err != nil {
		log.Error().Err(err).Msg("Error creating receiver")
		return nil, err
	}

	return h, nil
}

func newEventReceiver(pub *Publisher) func(context.Context, cloudevents.Event) protocol.Result {
	return func(ctx context.Context, e cloudevents.Event) protocol.Result {
		if e.Type() != EventTypeNewKubernetesVersionAvailable {
			log.Warn().Msgf("Received unexpected CloudEvent of type %q", e.Type())
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unexpected CloudEvent type %q", e.Type())
		}

		var ke NewKubernetesVersionAvailableEvent
		if err := e.DataAs(&ke); err != nil {
			log.Error().Err(err).Msg("Failed to deserialize NewKubernetesVersionAvailable data")
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "invalid NewKubernetesVersionAvailable data")
		}

		log.Info().Msgf("Received event with id %q", e.ID())
		if err := pub.Publish(ke); err != nil {
			log.Error().Err(err).Msg("Error publishing event")
		}

		return cloudevents.NewHTTPResult(http.StatusOK, "")
	}
}
