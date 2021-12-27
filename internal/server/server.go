package server

import (
	"log"

	"github.com/spf13/viper"
	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

type TimerecServer struct {
	logger        log.Logger
	stateProvider State
	backend       TimeService
	chat          NotificationService
}

type TimerecServerConfig struct {
	rocket providers.RocketChatConfig
}

type State interface {
	GetProfile() (api.Profile, error)
	UpdateProfile(api.Profile) (api.Profile, error)

	GetTemplates() ([]api.RecordTemplate, error)

	CreateWorkItem(api.WorkItem) (api.WorkItem, error)
	ListWorkItems() ([]api.WorkItem, error)
	GetWorkItem(api.WorkItem) (api.WorkItem, error)
	UpdateWorkItem(api.WorkItem) (api.WorkItem, error)
	DeleteWorkItem(api.WorkItem) (api.WorkItem, error)
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
	var chatService NotificationService

	var rocketConfig providers.RocketChatConfig
	err := viper.UnmarshalKey("rocketchat", &rocketConfig)
	if err == nil {
		rocket := providers.NewRocketChatMessenger(rocketConfig)
		chatService = &rocket
		// logger.Println("Using RocketChat NotificationService")
	} else {
		chatService = &providers.NoopProvider{}
		// logger.Println("Using Noop NotificationService")
	}

	return TimerecServer{
		logger:        *logger,
		stateProvider: fileBackend,
		backend:       fileBackend,
		chat:          chatService,
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
