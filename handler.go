package kubeup

import (
	"context"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/rs/zerolog/log"
)

const (
	EventTypeNewKubernetesVersionAvailable = "Microsoft.ContainerService.NewKubernetesVersionAvailable"
	AzureEventGridOrigin                   = "eventgrid.azure.net"
)

func NewCloudEventHandler(ctx context.Context, pub *Publisher) (http.Handler, error) {
	p, err := cloudevents.NewHTTP(cehttp.WithDefaultOptionsHandlerFunc(
		[]string{http.MethodOptions},
		cehttp.DefaultAllowedRate,
		[]string{AzureEventGridOrigin},
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
		vue := ResourceUpdateEvent{
			ResourceID:                         e.Source(),
			NewKubernetesVersionAvailableEvent: ke,
		}
		if err := pub.Publish(vue); err != nil {
			log.Error().Err(err).Msg("Error publishing event")
		}

		return cloudevents.NewHTTPResult(http.StatusOK, "")
	}
}
