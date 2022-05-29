package providers

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

type EventProvider struct {
	Sink string
}

func NewEventProvider(url string) (*EventProvider, error) {
	return &EventProvider{
		Sink: url,
	}, nil
}

func (prov *EventProvider) NotifyUser(ev cloudevents.Event) error {
	client, err := cloudevents.NewClientHTTP(http.WithTarget(prov.Sink))
	if err != nil {
		return err
	}

	result := client.Send(context.TODO(), ev)
	if cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}

	return nil
}
