package providers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/badkaktus/gorocket"
	"github.com/thomasbuchinger/timerec/api"
	"gopkg.in/yaml.v2"
)

type RocketChatMessenger struct {
	config RocketChatConfig
	logger log.Logger
	client gorocket.Client
}

type RocketChatConfig struct {
	Url, User, Token string
	Settings         struct {
		LoggerPrefix string `yaml:"loggerPrefix"`
		Channel      string
	}
	UserMappings     map[string]string `yaml:"userMappings"`
	MessageTemplates map[string]string `yaml:"messageTemplates"`
}

func NewRocketChatMessenger() RocketChatMessenger {
	config := loadRocketConfig()
	logger := log.New(os.Stdout, config.Settings.LoggerPrefix, 0)

	rocket := RocketChatMessenger{
		config: config,

		logger: *logger,
		client: *gorocket.NewClient(config.Url),
	}
	return rocket
}

func (rocket *RocketChatMessenger) EnsureLogin() error {
	me, err := rocket.client.Me()
	if err == nil && me.Success {
		rocket.logger.Println("Still logged in")
		// Still logged in, nothing to do
		return nil
	}

	rocket.logger.Println("Log into rocketchat")
	// Login again
	login := gorocket.LoginPayload{
		User:     rocket.config.User,
		Password: rocket.config.Token,
	}

	_, err = rocket.client.Login(&login)
	if err != nil {
		rocket.logger.Printf("Login failed: %+v", err)
		return err
	}
	return nil
}

func (rocket *RocketChatMessenger) NotifyUser(event api.Event) error {
	err := rocket.EnsureLogin()
	if err != nil {
		return err
	}

	message := gorocket.Message{
		Channel: rocket.config.Settings.Channel,
		Text:    rocket.generateMessage(event),
	}

	post_resp, err := rocket.client.PostMessage(&message)
	if err != nil && post_resp.Success {
		fmt.Printf("Error: %+v\n", err)
	}
	rocket.logger.Printf("Send message to '%s': %s\n", message.Channel, message.Text)
	return nil
}

func (rocket *RocketChatMessenger) generateMessage(event api.Event) string {
	var text bytes.Buffer
	templateString := rocket.config.MessageTemplates[event.Name]
	vars := struct {
		MappedUser, Target, User, EventName string
	}{
		MappedUser: rocket.config.UserMappings[event.User],
		Target:     event.Target,
		User:       event.User,
		EventName:  event.Name,
	}

	if templateString == "" {
		return event.Message
	}

	tmpl, err := template.New("RocketMessage").Parse(templateString)
	if err != nil {
		rocket.logger.Printf("failed to parse '%s'\n", event.Name)
		return event.Message
	}
	err = tmpl.Execute(&text, vars)
	if err != nil {
		rocket.logger.Printf("failed to execute template '%s'\n", event.Name)
		return event.Message
	}

	return text.String()
}

func loadRocketConfig() RocketChatConfig {
	var data RocketChatConfig

	content, err := os.ReadFile("cred_rocket.yaml")
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		fmt.Println(err)
	}
	return data

}
