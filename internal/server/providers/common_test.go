package providers_test

import (
	"testing"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

func TestAppendUser(t *testing.T) {
	users := []api.User{
		api.NewDefaultUser("testuser1"),
	}
	data := providers.StateV2{
		Users: users,
	}

	providers.CreateUser(&data, api.NewDefaultUser("testuser2"))

	if len(data.Users) != 2 {
		t.Log(data.Users)
		t.Fail()
	}
}
