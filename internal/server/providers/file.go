package providers

import (
	"fmt"
	"os"

	"github.com/thomasbuchinger/timerec/api"
	"gopkg.in/yaml.v2"
)

type FileProvider struct {
	Data FileData
}
type FileData struct {
	Profile   api.Profile
	Templates []api.RecordTemplate
	Tasks     map[string]api.Task
	Records   []api.Record
}

func (store *FileProvider) GetProfile() (api.Profile, error) {
	data := store.load()
	return data.Profile, nil
}

func (store *FileProvider) UpdateProfile(new api.Profile) (api.Profile, error) {
	data := store.load()
	data.Profile = new
	store.store(data)

	return data.Profile, nil
}
func (store *FileProvider) GetTemplates() ([]api.RecordTemplate, error) {
	data := store.load()

	return data.Templates, nil
}

func (store *FileProvider) CreateTask(t api.Task) (api.Task, error) {
	data := store.load()
	for name := range data.Tasks {
		if name == t.Name {
			return api.Task{}, fmt.Errorf("CONFLICT")
		}
	}
	data.Tasks[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) ListTasks() ([]api.Task, error) {
	data := store.load()
	taskList := []api.Task{}
	for _, task := range data.Tasks {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (store *FileProvider) GetTask(t api.Task) (api.Task, error) {
	data := store.load()
	for k, task := range data.Tasks {
		if k == t.Name {
			return task, nil
		}
	}
	return api.Task{}, fmt.Errorf("NOT_FOUND")
}

func (store *FileProvider) UpdateTask(t api.Task) (api.Task, error) {
	data := store.load()
	data.Tasks[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) DeleteTask(t api.Task) (api.Task, error) {
	data := store.load()
	for _, existing_task := range data.Tasks {
		if existing_task.Name == t.Name {
			delete(data.Tasks, existing_task.Name)
			store.store(data)
			return existing_task, nil
		}
	}
	return api.Task{}, fmt.Errorf("NOT_FOUND")
}

func (store *FileProvider) SaveRecord(rec api.Record) (api.Record, error) {
	data := store.load()
	data.Records = append(data.Records, rec)
	store.store(data)
	return rec, nil
}

func (store *FileProvider) load() FileData {
	var data FileData

	content, err := os.ReadFile("db.yaml")
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
	content, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile("db.yaml", content, os.ModeType)
	if err != nil {
		fmt.Println(err)
	}
}
