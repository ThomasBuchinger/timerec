package api

type Task struct {
	Id          string
	Name        string
	CustomerRef string

	Activities []TimeEntry
}

func (t Task) ConvertToRecords() []Record {
	var records []Record

	for _, activity := range t.Activities {
		records = append(records, Record{
			Id:          t.Id,
			Name:        t.Name,
			CustomerRef: t.CustomerRef,
			Description: activity.Comment,
			Start:       activity.Start,
			End:         activity.End,
		})
	}
	return records
}
