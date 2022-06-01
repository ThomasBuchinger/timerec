package providers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type WebhookProvider struct {
	Sink string
}

func NewWebhookProvider(url string) (*WebhookProvider, error) {
	return &WebhookProvider{
		Sink: url,
	}, nil
}

func (prov *WebhookProvider) NotifyUser(ev cloudevents.Event) error {
	jsonBytes, _ := ev.MarshalJSON()

	log.Printf("Sending Cloudevent to '%s'. Data: %s \n", prov.Sink, jsonBytes)
	resp, err := http.Post(prov.Sink, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Printf("Failed to send CloudEvent: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Received status: %s - %s", resp.Status, string(body))
		return fmt.Errorf("server error")
	}

	log.Println("Notification sent")
	return nil

	// client, err := cloudevents.NewClientHTTP(cloudevents.WithTarget(prov.Sink))
	// if err != nil {
	// 	return err
	// }

	// ctx := cloudevents.ContextWithTarget(context.Background(), prov.Sink)
	// result := client.Send(ctx, ev)
	// if cloudevents.IsUndelivered(result) {
	// 	log.Fatalf("failed to send, %v", result)
	// }

	// return nil
}
