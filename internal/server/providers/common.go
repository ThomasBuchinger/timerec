package providers

import (
	"github.com/thomasbuchinger/timerec/api"
)

type ProviderReturnType string

const (
	ProviderOk          ProviderReturnType = "OK"
	ProviderNotFound    ProviderReturnType = "NOT_FOUND"
	ProviderConflict    ProviderReturnType = "CONFLICT"
	ProviderForbidden   ProviderReturnType = "FORBIDDEN"
	ProviderServerError ProviderReturnType = "SERVER_ERROR"
)
const ScopeGlobal string = "global"

func (prov ProviderReturnType) Error() string {
	return string(prov)
}

type StateV2 struct {
	Partition string
	Users     []api.User
	Jobs      []api.Job
	Templates []api.RecordTemplate
	Records   []api.Record
}

func ListUsers(data *StateV2) ([]api.User, ProviderReturnType) {
	// Dummy for possible Filter Methods
	return data.Users, ProviderOk
}

func GetUser(data *StateV2, u api.User) (api.User, ProviderReturnType) {
	for _, user := range data.Users {
		if user.Name == u.Name {
			return user, ProviderOk
		}
	}
	return api.User{}, ProviderNotFound
}

func CreateUser(data *StateV2, new api.User) ProviderReturnType {
	_, err := GetUser(data, new)
	if err != ProviderNotFound {
		return ProviderConflict
	}
	data.Users = append(data.Users, new)
	return ProviderOk
}

func UpdateUser(data *StateV2, updated api.User) ProviderReturnType {
	for i, user := range data.Users {
		if user.Name == updated.Name {
			data.Users[i] = updated
			return ProviderOk
		}
	}
	return ProviderNotFound
}

func ListTemplates(data *StateV2) ([]api.RecordTemplate, ProviderReturnType) {
	return data.Templates, ProviderOk
}

func GetTemplate(data *StateV2, name string) (api.RecordTemplate, ProviderReturnType) {
	for _, tmpl := range data.Templates {
		if tmpl.TemplateName == name {
			return tmpl, ProviderOk
		}
	}
	return api.RecordTemplate{}, ProviderNotFound
}

func HasTemplate(data *StateV2, name string) (bool, ProviderReturnType) {
	_, ret := GetTemplate(data, name)
	return ret == ProviderOk, ret
}

func ListJobs(data *StateV2) ([]api.Job, ProviderReturnType) {
	return data.Jobs, ProviderOk
}

func GetJob(data *StateV2, j api.Job) (api.Job, ProviderReturnType) {
	for _, task := range data.Jobs {
		if task.Name == j.Name && task.Owner == j.Owner {
			return task, ProviderOk
		}
	}
	return api.Job{}, ProviderNotFound
}

func CreateJob(data *StateV2, new api.Job) ProviderReturnType {
	for _, j := range data.Jobs {
		if j.Name == new.Name {
			return ProviderConflict
		}
	}
	data.Jobs = append(data.Jobs, new)
	return ProviderOk
}

func UpdateJob(data *StateV2, updated api.Job) ProviderReturnType {
	for i, job := range data.Jobs {
		if job.Name == updated.Name {
			if job.Owner != updated.Owner {
				return ProviderForbidden
			}
			data.Jobs[i] = updated
			return ProviderOk
		}
	}
	return ProviderNotFound
}

func DeleteJob(data *StateV2, del api.Job) (api.Job, ProviderReturnType) {
	for i, job := range data.Jobs {
		if job.Name == del.Name {
			if job.Owner != del.Owner {
				return api.Job{}, ProviderForbidden
			}
			deleted := data.Jobs[i]
			updated_jobs := data.Jobs[:i]
			data.Jobs = append(updated_jobs, data.Jobs[i+1:]...)

			return deleted, ProviderOk
		}
	}
	return api.Job{}, ProviderNotFound
}

func SaveRecord(data *StateV2, rec api.Record) ProviderReturnType {
	data.Records = append(data.Records, rec)
	return ProviderOk
}
