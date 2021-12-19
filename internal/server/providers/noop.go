package providers

import (
	"fmt"

	"github.com/thomasbuchinger/timerec/api"
)

type NoopProvider struct{}

func (noop *NoopProvider) NotifyUser(event api.Event) error {
	fmt.Printf("Event: %s/%s: %s\n", event.Target, event.Name, event.Message)
	return nil
}
