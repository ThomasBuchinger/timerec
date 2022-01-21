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
	User      api.User
	Templates []api.RecordTemplate
	Tasks     map[string]api.Job
	Records   []api.Record
}

func (store *FileProvider) GetUser() (api.User, error) {
	data := store.load()
	return data.User, nil
}

func (store *FileProvider) UpdateUser(new api.User) (api.User, error) {
	data := store.load()
	data.User = new
	store.store(data)

	return data.User, nil
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
	return api.RecordTemplate{}, fmt.Errorf("not found")
}

func (store *FileProvider) CreateJob(t api.Job) (api.Job, error) {
	data := store.load()
	for name := range data.Tasks {
		if name == t.Name {
			return api.Job{}, fmt.Errorf("CONFLICT")
		}
	}
	data.Tasks[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) ListJobs() ([]api.Job, error) {
	data := store.load()
	taskList := []api.Job{}
	for _, task := range data.Tasks {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (store *FileProvider) GetJob(t api.Job) (api.Job, error) {
	data := store.load()
	for k, task := range data.Tasks {
		if k == t.Name {
			return task, nil
		}
	}
	return api.Job{}, fmt.Errorf("NOT_FOUND")
}

func (store *FileProvider) UpdateJob(t api.Job) (api.Job, error) {
	data := store.load()
	data.Tasks[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileProvider) DeleteJob(t api.Job) (api.Job, error) {
	data := store.load()
	for _, existing_task := range data.Tasks {
		if existing_task.Name == t.Name {
			delete(data.Tasks, existing_task.Name)
			store.store(data)
			return existing_task, nil
		}
	}
	return api.Job{}, fmt.Errorf("NOT_FOUND")
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
