package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	"go.uber.org/zap"
)

func NewTestServer(mem *providers.MemoryProvider) server.TimerecServer {
	logger, _ := zap.NewDevelopment()

	return server.TimerecServer{
		Logger:           *logger.Sugar(),
		StateProvider:    mem,
		TimeProvider:     mem,
		TemplateProvider: mem,
		ChatProvider:     mem,
	}
}

// Activity cannot be started, when another one is already active
func TestStartActivityIfAnotherActivityIsActive(t *testing.T) {
	mem := providers.MemoryProvider{User: api.User{Activity: api.Activity{
		ActivityName:    "exists",
		ActivityComment: "comment",
		ActivityStart:   time.Now(),
		ActivityTimer:   time.Now(),
	}}}
	mgr := NewTestServer(&mem)

	response, err := mgr.StartActivity(context.TODO(), server.StartActivityParams{
		UserName:         "me",
		ActivityName:     "new",
		StartDuration:    time.Duration(0),
		EstimateDuration: time.Duration(0),
	})

	if mem.User.Activity.ActivityName != "exists" {
		t.Logf("ActivityName updated, despite error. got %s, expected exists", mem.User.Activity.ActivityName)
		t.Fail()
	}
	if response.Success != false {
		t.Logf("Response.Success was %t, expected false", response.Success)
		t.Fail()
	}
	if err == nil {
		t.Logf("Expected error, got nil: %s", err.Error())
		t.Fail()
	}
}

func TestStartActivityWorks(t *testing.T) {
	mem := providers.MemoryProvider{}
	mgr := NewTestServer(&mem)

	response, _ := mgr.StartActivity(context.TODO(), server.StartActivityParams{
		UserName:         "me",
		ActivityName:     "new",
		StartDuration:    time.Duration(0),
		EstimateDuration: time.Duration(0),
	})
	if mem.User.Activity.ActivityName != "new" {
		t.Logf("ActivityName not updated. got %s", mem.User.Activity.ActivityName)
		t.Fail()
	}
	if response.Activity.ActivityName != "new" {
		t.Logf("ActivityName not in response. got %s", response.Activity.ActivityName)
		t.Fail()
	}
	if response.Success != true {
		t.Logf("Response.Success was %t, expected true", response.Success)
		t.Fail()
	}
}

func TestExtendActivityWorks(t *testing.T) {
	mem := providers.MemoryProvider{}
	dur, _ := time.ParseDuration("1h")
	mgr := NewTestServer(&mem)

	mgr.StartActivity(context.TODO(), server.StartActivityParams{
		UserName:         "me",
		ActivityName:     "new",
		StartDuration:    dur,
		EstimateDuration: dur,
	})
	ts1 := mem.User.Activity.ActivityTimer
	mgr.ExtendActivity(context.TODO(), server.ExtendActivityParams{
		UserName: "me",
		Estimate: dur + dur,
	})
	if ts1 == mem.User.Activity.ActivityTimer {
		t.Fatal("timestamp not updated")
	}
}

func TestFinishActivityWorks(t *testing.T) {
	mem := providers.MemoryProvider{Jobs: map[string]api.Job{}}
	dur, _ := time.ParseDuration("30m")
	mgr := NewTestServer(&mem)

	mgr.CreateJobIfMissing(context.TODO(), server.SearchJobParams{
		Name:          "testwork",
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})
	mgr.StartActivity(context.TODO(), server.StartActivityParams{
		UserName:         "me",
		ActivityName:     "test",
		StartDuration:    dur,
		EstimateDuration: dur,
	})

	act := mem.User.Activity
	if err := act.CheckActivityActive(); err != nil {
		t.Fatal("StartActivity did not work", act)
	}

	res, fin_err := mgr.FinishActivity(context.TODO(), server.FinishActivityParams{
		UserName:     "me",
		JobName:      "testwork",
		ActivityName: "test",
		EndDuration:  dur,
	})
	act = mem.User.Activity
	if fin_err != nil {
		t.Logf("Expecting no Error, got %v", fin_err)
		t.Fail()
	}
	if err := act.CheckNoActivityActive(); err != nil {
		t.Log("FinishActivity did not clear Activity from Profile")
		t.Fail()
	}
	if !res.Success {
		t.Logf("FinishActivity returned %t", res.Success)
		t.Fail()
	}
	work, ok := mem.Jobs["testwork"]
	if !ok || len(work.Activities) != 1 {
		t.Log("Job seems to be empty", work)
		t.Fail()
	}
}

// Activities cannot be finish without a Job
func TestFinishActivityFailsWithoutJob(t *testing.T) {
	mem := providers.MemoryProvider{Jobs: map[string]api.Job{}}
	dur, _ := time.ParseDuration("30m")
	mgr := NewTestServer(&mem)

	mgr.StartActivity(context.TODO(), server.StartActivityParams{
		UserName:         "me",
		ActivityName:     "test",
		StartDuration:    dur,
		EstimateDuration: dur,
	})

	act := mem.User.Activity
	if err := act.CheckActivityActive(); err != nil {
		t.Fatal("StartActivity did not work", act)
	}

	res, err := mgr.FinishActivity(context.TODO(), server.FinishActivityParams{
		UserName:     "me",
		JobName:      "testwork",
		ActivityName: "test",
		EndDuration:  dur,
	})
	act = mem.User.Activity
	if err == nil {
		t.Log("expected no error, got nothing")
		t.Fail()
	}
	if res.Success {
		t.Fatalf("FinishActivity returned. got %t expected %t", res.Success, false)
	}
	if err := act.CheckActivityActive(); err != nil {
		t.Fatalf("Operation failed, but Activity still cleared. CheckActivityActive() returned %v", err)
	}
}

func TestJobIfMissingWorks(t *testing.T) {
	mem := providers.MemoryProvider{Jobs: map[string]api.Job{}}
	mgr := NewTestServer(&mem)

	res, err := mgr.CreateJobIfMissing(context.TODO(), server.SearchJobParams{
		Name:          "testwork",
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})

	if err != nil {
		t.Logf("expected not error, got: %v", err)
	}
	if !res.Success {
		t.Fatalf("response was not successful. got %t expected %t", res.Success, true)
	}
	if !res.Created {
		t.Fatalf("response reported no Job created. got %t expected %t", res.Created, true)
	}
	if len(mem.Jobs) != 1 {
		t.Fatalf("incorrect number of Jobs: got %d expected %d", len(mem.Jobs), 1)
	}
}

func TestCreateJobIfMissingIsIdempotent(t *testing.T) {
	mem := providers.MemoryProvider{Jobs: map[string]api.Job{}}
	mgr := NewTestServer(&mem)

	mgr.CreateJobIfMissing(context.TODO(), server.SearchJobParams{
		Name:          "testwork",
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})
	res2, err := mgr.CreateJobIfMissing(context.TODO(), server.SearchJobParams{
		Name:          "testwork",
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})

	if err != nil {
		t.Logf("expected no error, got %v", err)
	}
	if !res2.Success {
		t.Fatalf("response was not successful. got %t expected %t", res2.Success, true)
	}
	if res2.Created {
		t.Fatalf("response reported no Job created. got %t expected %t", res2.Created, false)
	}
	if len(mem.Jobs) != 1 {
		t.Fatalf("incorrect number of Jobs: got %d expected %d", len(mem.Jobs), 1)
	}
}
