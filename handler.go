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
	AzureEventGridOrigin = "eventgrid.azure.net"
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

func newEventReceiver(p *Publisher) func(context.Context, cloudevents.Event) protocol.Result {
	return func(ctx context.Context, e cloudevents.Event) protocol.Result {
		log.Info().Msgf("Received event with id %q", e.ID())
		switch e.Type() {
		case EventNewKubernetesVersionAvailable:
			return publishEvent[ContainerServiceNewKubernetesVersionAvailableEvent](e, p, "new-kubernetes-version.gohtml")
		case EventClusterSupportEnding:
			return publishEvent[ContainerServiceClusterSupportEndingEvent](e, p, "cluster-support-ending.gohtml")
		case EventClusterSupportEnded:
			return publishEvent[ContainerServiceClusterSupportEndedEvent](e, p, "cluster-support-ended.gohtml")
		case EventNodePoolRollingStarted:
			return publishEvent[ContainerServiceNodePoolRollingStartedEvent](e, p, "nodepool-rolling-started.gohtml")
		case EventNodePoolRollingSucceeded:
			return publishEvent[ContainerServiceNodePoolRollingSucceededEvent](e, p, "nodepool-rolling-succeeded.gohtml")
		case EventNodePoolRollingFailed:
			return publishEvent[ContainerServiceNodePoolRollingFailedEvent](e, p, "nodepool-rolling-failed.gohtml")
		default:
			log.Warn().Msgf("Received unexpected CloudEvent of type %q", e.Type())
			return cloudevents.NewHTTPResult(http.StatusBadRequest, "unexpected CloudEvent type %q", e.Type())
		}
	}
}

func publishEvent[T ContainerServiceEvent](e cloudevents.Event, p *Publisher, filename string) protocol.Result {
	ce, err := unmarshal[T](e)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to deserialize %s data", e.Type())
		return cloudevents.NewHTTPResult(http.StatusBadRequest, "invalid %s data", e.Type())
	}
	mb := NewMessageBuilder[T](filename)
	msg, err := mb.Build(ce, e.Source())
	if err != nil {
		log.Error().Err(err).Msg("Error building message")
		return cloudevents.NewHTTPResult(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	if err := p.Publish(msg); err != nil {
		log.Error().Err(err).Msg("Error publishing event")
	}
	return cloudevents.NewHTTPResult(http.StatusOK, "")
}

func unmarshal[T ContainerServiceEvent](e cloudevents.Event) (T, error) {
	var data T
	if err := e.DataAs(&data); err != nil {
		return data, err
	}
	return data, nil
}
