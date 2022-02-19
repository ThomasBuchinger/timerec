package providers

import (
	"errors"
	"fmt"
	"os"

	"github.com/thomasbuchinger/timerec/api"
	"gopkg.in/yaml.v2"
)

type FileProvider struct {
	Path string
	Data FileData
}
type FileData struct {
	User      map[string]api.User
	Templates []api.RecordTemplate
	Jobs      map[string]api.Job
	Records   []api.Record
}

func NewMemoryProvider() *FileProvider {
	mem := FileProvider{}
	mem.Data = FileData{
		User:      map[string]api.User{},
		Templates: []api.RecordTemplate{},
		Jobs:      map[string]api.Job{},
		Records:   []api.Record{},
	}
	return &mem
}

func (store *FileProvider) ListUsers() ([]api.User, error) {
	data := store.load()
	var ret []api.User

	for _, user := range data.User {
		ret = append(ret, user)
	}
	return ret, nil
}

func (store *FileProvider) GetUser(u api.User) (api.User, error) {
	data := store.load()
	for username, user := range data.User {
		if username == u.Name {
			return user, nil
		}
	}
	return api.User{}, errors.New(string(ProviderErrorNotFound))
}

func (store *FileProvider) CreateUser(new api.User) (api.User, error) {
	data := store.load()
	for username := range data.User {
		if username == new.Name {
			return api.User{}, errors.New(string(ProviderErrorConflict))
		}
	}

	data.User[new.Name] = new
	store.store(data)
	return new, nil
}

func (store *FileProvider) UpdateUser(new api.User) (api.User, error) {
	data := store.load()
	for username, _ := range data.User {
		if username == new.Name {
			data.User[new.Name] = new
			store.store(data)

			return new, nil
		}
	}

	return api.User{}, errors.New(string(ProviderErrorNotFound))
}

func (store *FileProvider) GetTemplates() ([]api.RecordTemplate, error) {
	data := store.load()
	return data.Templates, nil
}

func (store *FileProvider) HasTemplate(name string) (bool, error) {
	data := store.load()
	for _, tmpl := range data.Templates {
		if tmpl.TemplateName == name {
			return true, nil
		}
	}
	return false, nil
}

func (store *FileProvider) GetTemplate(name string) (api.RecordTemplate, error) {
	data := store.load()
	for _, tmpl := range data.Templates {
		if tmpl.TemplateName == name {
			return tmpl, nil
		}
	}
	return api.RecordTemplate{}, errors.New(string(ProviderErrorNotFound))
}

func (store *FileProvider) CreateJob(t api.Job) (api.Job, error) {
	data := store.load()
	for name := range data.Jobs {
		if name == t.Name {
			return api.Job{}, errors.New(string(ProviderErrorConflict))
		}
	}
	data.Jobs[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) ListJobs() ([]api.Job, error) {
	data := store.load()
	taskList := []api.Job{}
	for _, task := range data.Jobs {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (store *FileProvider) GetJob(t api.Job) (api.Job, error) {
	data := store.load()
	for _, task := range data.Jobs {
		if task.Name == t.Name && task.Owner == t.Owner {
			return task, nil
		}
	}
	return api.Job{}, errors.New(string(ProviderErrorNotFound))
}

func (store *FileProvider) UpdateJob(t api.Job) (api.Job, error) {
	data := store.load()
	existing, ok := data.Jobs[t.Name]
	if !ok {
		return api.Job{}, errors.New(string(ProviderErrorNotFound))
	}
	if existing.Owner != t.Owner {
		return api.Job{}, errors.New(string(ProviderErrorForbidden))
	}

	data.Jobs[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) DeleteJob(t api.Job) (api.Job, error) {
	data := store.load()
	for _, existing_task := range data.Jobs {
		if existing_task.Name == t.Name {
			if existing_task.Owner != t.Owner {
				return api.Job{}, errors.New(string(ProviderErrorForbidden))
			}

			delete(data.Jobs, existing_task.Name)
			store.store(data)
			return existing_task, nil
		}
	}
	return api.Job{}, errors.New(string(ProviderErrorNotFound))
}

func (store *FileProvider) SaveRecord(rec api.Record) (api.Record, error) {
	data := store.load()
	data.Records = append(data.Records, rec)
	store.store(data)
	return rec, nil
}

func (store *FileProvider) NotifyUser(event api.Event) error {
	fmt.Printf("Event: %s/%s/%s: %s\n", event.User, event.Target, event.Name, event.Message)
	return nil
}

func (store *FileProvider) load() FileData {
	if store.Path == "" {
		return store.Data
	}

	var data FileData
	content, err := os.ReadFile(store.Path)
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		fmt.Println(err)
	}
	return data

}
func (store *FileProvider) store(data FileData) {
	if store.Path == "" {
		store.Data = data
		return
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(store.Path, content, os.ModeType)
	if err != nil {
		fmt.Println(err)
	}
}
