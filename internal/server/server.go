package server

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	"go.uber.org/zap"
)

type TimerecServer struct {
	logger           log.Logger
	loggerv2         zap.SugaredLogger
	StateProvider    State
	TemplateProvider TemplateService
	TimeProvider     TimeService
	ChatProvider     NotificationService
}

type TimerecServerConfig struct {
	Settings struct {
		RoundTo time.Duration
	}
	rocket providers.RocketChatConfig
}

type State interface {
	GetUser() (api.User, error)
	UpdateUser(api.User) (api.User, error)

	CreateWorkItem(api.WorkItem) (api.WorkItem, error)
	ListWorkItems() ([]api.WorkItem, error)
	GetWorkItem(api.WorkItem) (api.WorkItem, error)
	UpdateWorkItem(api.WorkItem) (api.WorkItem, error)
	DeleteWorkItem(api.WorkItem) (api.WorkItem, error)
}

type TemplateService interface {
	GetTemplates() ([]api.RecordTemplate, error)
	HasTemplate(string) (bool, error)
	GetTemplate(string) (api.RecordTemplate, error)
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
	loggerv2, err := zap.NewProduction()
	fileBackend := &providers.FileProvider{}

	var chatService NotificationService
	var rocketConfig providers.RocketChatConfig
	err = viper.UnmarshalKey("rocketchat", &rocketConfig)
	if err == nil {
		rocket := providers.NewRocketChatMessenger(rocketConfig)
		chatService = &rocket
		// logger.Println("Using RocketChat NotificationService")
	} else {
		chatService = &providers.NoopProvider{}
		// logger.Println("Using Noop NotificationService")
	}

	return TimerecServer{
		logger:           *logger,
		loggerv2:         *loggerv2.Sugar(),
		StateProvider:    fileBackend,
		TimeProvider:     fileBackend,
		TemplateProvider: fileBackend,
		ChatProvider:     chatService,
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
