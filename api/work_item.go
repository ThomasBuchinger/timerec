package api

import (
	"fmt"
	"time"
)

type WorkItem struct {
	Name           string    `yaml:"task_name"`
	CreatedAt      time.Time `yaml:"created,omitempty"`
	RecordTemplate `yaml:",inline"`

	Activities []TimeEntry `yaml:"activities"`
}

type RecordTemplate struct {
	TemplateName string `yaml:"template_name"`
	Project      string `yaml:"project,omitempty"`
	Task         string `yaml:"task,omitempty"`
	Title        string `yaml:"title,omitempty"`
	Description  string `yaml:"description,omitempty"`
}

type TimeEntry struct {
	Comment string    `yaml:"comment,omitempty"`
	Start   time.Time `yaml:"start,omitempty"`
	End     time.Time `yaml:"end,omitempty"`
}

func NewWorkItem(name string) WorkItem {
	return WorkItem{
		Name:      name,
		CreatedAt: time.Now(),
	}
}

func (t *WorkItem) Validate() error {
	var missingCommentsInActivities bool = false
	for _, act := range t.Activities {
		if act.Comment == "" {
			missingCommentsInActivities = true
		}
	}

	if t.RecordTemplate.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if t.RecordTemplate.Description == "" && missingCommentsInActivities {
		return fmt.Errorf("description cannot be empty")
	}
	if t.RecordTemplate.Project == "" {
		return fmt.Errorf("project cannot be empty")
	}
	if t.RecordTemplate.Task == "" {
		return fmt.Errorf("task cannot be empty")
	}
	return nil
}

func (t *WorkItem) ConvertToRecords() []Record {
	var records []Record

	for _, activity := range t.Activities {
		var desc string
		if activity.Comment == "" {
			desc = t.RecordTemplate.Description
		} else {
			desc = t.RecordTemplate.Description + "\n" + activity.Comment
		}

		records = append(records, Record{
			Title:       t.Title,
			Description: desc,
			Project:     t.RecordTemplate.Project,
			Task:        t.RecordTemplate.Task,
			Start:       activity.Start,
			End:         activity.End,
		})
	}
	return records
}

func (t *WorkItem) Update(new WorkItem) error {
	if new.RecordTemplate.Title != "" {
		t.RecordTemplate.Title = new.RecordTemplate.Title
	}
	if new.RecordTemplate.Description != "" {
		t.RecordTemplate.Description = new.RecordTemplate.Description
	}
	if new.RecordTemplate.Project != "" {
		t.RecordTemplate.Project = new.RecordTemplate.Project
	}
	if new.RecordTemplate.Task != "" {
		t.RecordTemplate.Task = new.RecordTemplate.Task
	}

	// Update Activities
	for _, newAct := range new.Activities {
		err := t.AddActivity(newAct)
		if err != nil {
			return err
		}
	}
	return nil
}
func (t *WorkItem) AddActivity(new_activity TimeEntry) error {
	for _, existing_activity := range t.Activities {
		if new_activity.Start == existing_activity.Start && new_activity.End == existing_activity.End {
			if new_activity.Comment != "" {
				existing_activity.Comment = new_activity.Comment
			}
			return nil
		}
	}
	t.Activities = append(t.Activities, new_activity)
	return nil
}
