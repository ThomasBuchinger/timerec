package client

import "github.com/thomasbuchinger/timerec/api"

func EditRecordsPreSendHook(rec []api.Record) []api.Record {
	return []api.Record{}
}
