package providers

import (
	"fmt"

	"github.com/thomasbuchinger/timerec/api"
)

type MemoryProvider struct {
	User      api.User
	Templates []api.RecordTemplate
	Jobs      map[string]api.Job
	Records   []api.Record
}

func (mem *MemoryProvider) GetUser() (api.User, error) {
	return mem.User, nil
}

func (mem *MemoryProvider) UpdateUser(new api.User) (api.User, error) {
	mem.User = new
	return mem.User, nil
}

func (mem *MemoryProvider) GetTemplates() ([]api.RecordTemplate, error) {
	return mem.Templates, nil
}

func (mem *MemoryProvider) HasTemplate(name string) (bool, error) {
	for _, tmpl := range mem.Templates {
		if tmpl.TemplateName == name {
			return true, nil
		}
	}
	return false, nil
}

func (mem *MemoryProvider) GetTemplate(name string) (api.RecordTemplate, error) {
	for _, tmpl := range mem.Templates {
		if tmpl.TemplateName == name {
			return tmpl, nil
		}
	}
	return api.RecordTemplate{}, fmt.Errorf("not found")
}

func (mem *MemoryProvider) CreateJob(t api.Job) (api.Job, error) {
	for name := range mem.Jobs {
		if name == t.Name {
			return api.Job{}, fmt.Errorf("CONFLICT")
		}
	}
	mem.Jobs[t.Name] = t
	return t, nil
}

func (mem *MemoryProvider) ListJobs() ([]api.Job, error) {
	taskList := []api.Job{}
	for _, task := range mem.Jobs {
		taskList = append(taskList, task)
	}

	return taskList, nil
}

func (mem *MemoryProvider) GetJob(t api.Job) (api.Job, error) {
	for k, task := range mem.Jobs {
		if k == t.Name {
			return task, nil
		}
	}
	return api.Job{}, fmt.Errorf("NOT_FOUND")
}

func (mem *MemoryProvider) UpdateJob(t api.Job) (api.Job, error) {
	mem.Jobs[t.Name] = t
	return t, nil
}

func (mem *MemoryProvider) DeleteJob(t api.Job) (api.Job, error) {
	for _, existing_task := range mem.Jobs {
		if existing_task.Name == t.Name {
			delete(mem.Jobs, existing_task.Name)
			return existing_task, nil
		}
	}
	return api.Job{}, fmt.Errorf("NOT_FOUND")
}

func (mem *MemoryProvider) SaveRecord(rec api.Record) (api.Record, error) {
	mem.Records = append(mem.Records, rec)
	return rec, nil
}

func (mem *MemoryProvider) NotifyUser(event api.Event) error {
	fmt.Printf("Event: %s/%s: %s\n", event.Target, event.Name, event.Message)
	return nil
}
