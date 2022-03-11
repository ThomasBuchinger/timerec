package api

import (
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Event struct {
	Name    string
	Message string
	Target  string
	User    string
}

type EventType string

const (
	EventTypeTimerExpired EventType = "TIMER_EXPIRED"
	EventTypeNoEntryAlarm EventType = "NO_ENTRY_ALARM"
)

type Message struct {
	User    string `json:"user,omitempty"`
	Message string `json:"message"`
}

func MakeMessageEvent(name EventType, message, target, user string) cloudevents.Event {
	ev := cloudevents.NewEvent()
	ev.SetSpecVersion(cloudevents.VersionV1)
	ev.SetType("sh.buc.ChatMessage.Send")
	ev.SetSource("timerec")
	ev.SetSubject(user)
	ev.SetID(uuid.New().String())
	ev.SetTime(time.Now())
	data := Message{
		User:    fmt.Sprintf("@%s", user),
		Message: message,
	}
	ev.SetData("application/json", &data)

	return ev
}
