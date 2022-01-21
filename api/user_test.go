package api_test

import (
	"testing"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

func TestCheckActivityActive(t *testing.T) {
	testCases := []struct {
		desc string
		api.Activity
		expected bool
	}{
		{desc: "name-only", expected: true, Activity: api.Activity{ActivityName: "name-only"}},
		{desc: "full", expected: true, Activity: api.Activity{ActivityName: "full", ActivityStart: time.Now(), ActivityTimer: time.Now()}},
		{desc: "empty", expected: false, Activity: api.Activity{}},
		{desc: "name-empty-times-set", expected: false, Activity: api.Activity{ActivityStart: time.Now(), ActivityTimer: time.Now()}},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.Activity.CheckActivityActive()
			actual := err == nil
			if actual != tC.expected {
				t.Logf("Expected err == nil to be %t, got %t. ", tC.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestCheckNoActivityActive(t *testing.T) {
	testCases := []struct {
		desc string
		api.Activity
		expected bool
	}{
		{desc: "name-only", expected: false, Activity: api.Activity{ActivityName: "name-only"}},
		{desc: "full", expected: false, Activity: api.Activity{ActivityName: "full", ActivityStart: time.Now(), ActivityTimer: time.Now()}},
		{desc: "empty", expected: true, Activity: api.Activity{}},
		{desc: "name-empty-times-set", expected: true, Activity: api.Activity{ActivityStart: time.Now(), ActivityTimer: time.Now()}},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.Activity.CheckNoActivityActive()
			actual := err == nil
			if actual != tC.expected {
				t.Logf("Expected err == nil to be %t, got %t. ", tC.expected, actual)
				t.Fail()
			}
		})
	}
}

func TestSetActivity(t *testing.T) {
	user := api.User{}

	err := user.Activity.CheckNoActivityActive()
	if err != nil {
		t.Fatal("Empty User does not pass CheckNoActivityActive")
	}
	user.SetActivity("my-activity", "", time.Now(), time.Now())
	err = user.Activity.CheckActivityActive()
	if err != nil {
		t.Fatalf("Set activity not recognized by CheckActivityActive: %s", err.Error())
	}
}

func TestClearActivity(t *testing.T) {
	user := api.User{}

	err := user.Activity.CheckNoActivityActive()
	if err != nil {
		t.Fatal("Empty User does not pass CheckNoActivityActive")
	}
	user.SetActivity("my-activity", "", time.Now(), time.Now())
	err = user.Activity.CheckActivityActive()
	if err != nil {
		t.Fatalf("Set activity not recognized by CheckActivityActive: %s", err.Error())
	}
	user.ClearActivity()
	err = user.Activity.CheckNoActivityActive()
	if err != nil {
		t.Fatalf("Activity still active after clear: %s", err.Error())
	}
}
