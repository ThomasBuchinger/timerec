package server

import (
	"fmt"
	"os"

	"github.com/thomasbuchinger/timerec/api"
	"gopkg.in/yaml.v2"
)

type FileStorage struct {
	Data FileData
}
type FileData struct {
	Profile api.Profile
	Tasks   map[string]api.Task
	Records []api.Record
}

func (store *FileStorage) CreateTask(t api.Task) (api.Task, error) {
	data := store.load()
	data.Tasks[t.Name] = t
	store.store(data)
	return t, nil
}

func (store *FileStorage) GetTasks() ([]api.Task, error) {
	data := store.load()
	taskList := make([]api.Task, len(data.Tasks))
	for _, task := range data.Tasks {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (store *FileStorage) AddRecord(rec api.Record) (api.Record, error) {
	data := store.load()
	data.Records = append(data.Records, rec)
	store.store(data)
	return rec, nil
}

func (store *FileStorage) load() FileData {
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
func (store *FileStorage) store(data FileData) {
	content, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile("db.yaml", content, os.ModeType)
	if err != nil {
		fmt.Println(err)
	}
}
