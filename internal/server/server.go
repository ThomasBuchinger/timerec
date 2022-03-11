package server

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/spf13/viper"
	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	"go.uber.org/zap"
)

type TimerecServer struct {
	Logger zap.SugaredLogger

	StateProvider State
	TimeProvider  TimeService
	ChatProvider  NotificationService
}

type TimerecServerConfig struct {
	File struct {
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
	} `json:"file,omitempty"`
	Kubernetes struct {
		Enabled    bool   `json:"enabled"`
		KubeConfig string `json:"kube_config,omitempty"`
	} `json:"kubernetes,omitempty"`
	Clockodo struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"clockodo,omitempty"`
	RocketChatBridge struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"rocket_chat_bridge,omitempty"`
}

type State interface {
	Refresh(string) (providers.StateV2, error)
	Save(string, providers.StateV2) error
}

type TimeService interface {
	SaveRecord(api.Record) (api.Record, error)
}
type NotificationService interface {
	// Different Services might have vastly different ideas how messages should look like
	// Events aim to be a very generic interface
	NotifyUser(cloudevents.Event) error
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

func (mgr *TimerecServer) MakeNewResponseError(t ResponseErrorType, err error, message string, values ...interface{}) ResponseError {
	respErr := ResponseError{
		Type:    t,
		Message: fmt.Sprintf(message, values...),
		Cause:   err,
	}
	mgr.Logger.Error(respErr)
	return respErr
}

const (
	BadRequest      ResponseErrorType = "BAD_REQUEST"
	ValidationError ResponseErrorType = "VALIDATION_FAILED"
	ProviderError   ResponseErrorType = "BACKEND_ERROR"
	ServerError     ResponseErrorType = "SERVER_ERROR"
)

func NewServer() TimerecServer {
	// logger, _ := zap.NewProduction()
	logger, _ := zap.NewDevelopment()
	defaultProvider := providers.NewMemoryProvider()
	server := TimerecServer{
		Logger:        *logger.Sugar(),
		StateProvider: defaultProvider,
		TimeProvider:  defaultProvider,
		ChatProvider:  defaultProvider,
	}

	var settings TimerecServerConfig
	err := viper.Unmarshal(&settings)
	if err != nil {
		logger.Warn(fmt.Sprintf("Config File invalid: %v", err))
	}

	// Configure File Provider
	if settings.File.Enabled {
		fileProvider := providers.NewFileProvider(settings.File.Path)
		server.StateProvider = fileProvider
		logger.Sugar().Debug("Using State: File")

		server.TimeProvider = fileProvider
		logger.Sugar().Debug("Using TimeService: File")
	}

	// Configure Kubernetes Provider
	if settings.Kubernetes.Enabled {
		kubernetesProvider, err := providers.NewKubernetesProvider(server.Logger, viper.GetString("kubernetes.kubeconfig"))
		if err != nil {
			panic(err)
		}
		server.StateProvider = kubernetesProvider
		logger.Sugar().Debug("Using State: Kubernetes")

		server.TimeProvider = kubernetesProvider
		logger.Sugar().Debug("Using TimeService: Kubernetes")
	}

	// Configure RocketChatBridge Provider
	if settings.RocketChatBridge.Enabled {
		webhookProvider, _ := providers.NewEventProvider(viper.GetString("rocket_chat_bridge.url"))
		server.ChatProvider = webhookProvider
		logger.Sugar().Debug("Using Chat: RocketChatBridge")
	}

	return server
}
