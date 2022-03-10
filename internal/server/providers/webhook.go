package providers

type EventProvider struct {
	Sink string
}

func NewEventProvider(url string) (*EventProvider, error) {
	return &EventProvider{
		Sink: url,
	}, nil
}
