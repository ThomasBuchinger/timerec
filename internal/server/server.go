package server

import (
	"log"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

type TimerecServer struct {
	logger        log.Logger
	stateProvider State
	backend       TimeService
	chat          NotificationService
}

type State interface {
	GetProfile() (api.Profile, error)
	UpdateProfile(api.Profile) (api.Profile, error)

	GetTemplates() ([]api.RecordTemplate, error)

	CreateTask(api.Task) (api.Task, error)
	ListTasks() ([]api.Task, error)
	GetTask(api.Task) (api.Task, error)
	UpdateTask(api.Task) (api.Task, error)
	DeleteTask(api.Task) (api.Task, error)
}

type TimeService interface {
	SaveRecord(api.Record) (api.Record, error)
}
type NotificationService interface {
	// Different Services might have vastly different ideas how messages should look like
	// Events aim to be a very generic interface
	NotifyUser(api.Event) error
}

func NewServer() TimerecServer {
	logger := log.Default()
	fileBackend := &providers.FileProvider{}
	// noop := &providers.NoopProvider{}
	rocket := providers.NewRocketChatMessenger()

	return TimerecServer{
		logger:        *logger,
		stateProvider: fileBackend,
		backend:       fileBackend,
		chat:          &rocket,
	}
}

func MakeEvent(name, message, target, user string) api.Event {
	return api.Event{
		Name:    name,
		Message: message,
		Target:  target,
		User:    "me",
	}
}
