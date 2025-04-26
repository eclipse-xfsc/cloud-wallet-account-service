package connection

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
	"gitlab.eclipse.org/eclipse/xfsc/libraries/messaging/cloudeventprovider"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/common"
	"github.com/eclipse-xfsc/cloud-wallet-account-service/internal/config"
)

var logger = common.GetLogger()

func CloudEventsConnectionSubscribe(topic string, handler func(e event.Event)) (*cloudeventprovider.CloudEventProviderClient, func() error, error) {
	// client, err := cloudeventprovider.NewClient(cloudeventprovider.Sub, topic)

	client, err := cloudeventprovider.New(cloudeventprovider.Config{
		Protocol: cloudeventprovider.ProtocolTypeNats,
		Settings: cloudeventprovider.NatsConfig{
			Url:        config.ServerConfiguration.Nats.Url,
			QueueGroup: config.ServerConfiguration.Nats.QueueGroup,
		},
	}, cloudeventprovider.ConnectionTypeSub, topic)

	if err != nil {
		logger.Error(err, "error during processing message")
		return nil, nil, err
	} else {
		logger.Info(fmt.Sprintf("cloudEvents can be received over topic: %s", topic))
	}
	return client, func() error {
		return client.SubCtx(context.Background(), handler)
	}, nil
}

func CloudEventsConnectionPublish(topic string, e event.Event) (*cloudeventprovider.CloudEventProviderClient, func() error, error) {
	client, err := cloudeventprovider.New(cloudeventprovider.Config{
		Protocol: cloudeventprovider.ProtocolTypeNats,
		Settings: cloudeventprovider.NatsConfig{
			Url:        config.ServerConfiguration.Nats.Url,
			QueueGroup: config.ServerConfiguration.Nats.QueueGroup,
		},
	}, cloudeventprovider.ConnectionTypePub, topic)

	if err != nil {
		logger.Error(err, "error during processing message")
		return nil, nil, err
	} else {
		logger.Info(fmt.Sprintf("cloudEvents can be published to topic: %s", topic))
	}
	return client, func() error {
		return client.PubCtx(context.TODO(), e)
	}, nil
}
