package server

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	"go.uber.org/zap"
)

type TimerecServer struct {
	Logger           zap.SugaredLogger
	StateProvider    State
	TemplateProvider TemplateService
	TimeProvider     TimeService
	ChatProvider     NotificationService
}

type TimerecServerConfig struct {
	Settings struct {
		RoundTo time.Duration
	}
}

type State interface {
	GetUser() (api.User, error)
	UpdateUser(api.User) (api.User, error)

	CreateJob(api.Job) (api.Job, error)
	ListJobs() ([]api.Job, error)
	GetJob(api.Job) (api.Job, error)
	UpdateJob(api.Job) (api.Job, error)
	DeleteJob(api.Job) (api.Job, error)
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

type ResponseError struct {
	Type    ResponseErrorType
	Message string
	Cause   error
}
type ResponseErrorType string

func (r ResponseError) Error() string {
	return fmt.Sprintf("%s: %v", r.Type, r.Cause)
}

const (
	BadRequest      ResponseErrorType = "BAD_REQUEST"
	ValidationError ResponseErrorType = "VALIDATION_FAILED"
	ProviderError   ResponseErrorType = "BACKEND_ERROR"
	ServerError     ResponseErrorType = "SERVER_ERROR"
)

func NewServer() TimerecServer {
	logger, _ := zap.NewProduction()
	fileBackend := &providers.FileProvider{}

	var chatService NotificationService
	var rocketConfig providers.RocketChatConfig
	err := viper.UnmarshalKey("rocketchat", &rocketConfig)
	if err == nil {
		rocket := providers.NewRocketChatMessenger(rocketConfig)
		chatService = &rocket
		logger.Sugar().Debug("Using RocketChat NotificationService")
	} else {
		chatService = &providers.MemoryProvider{}
		logger.Sugar().Debug("Using Noop NotificationService")
	}

	return TimerecServer{
		Logger:           *logger.Sugar(),
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
