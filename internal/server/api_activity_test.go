package server_test

import (
	"github.com/thomasbuchinger/timerec/internal/server"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	// "github.com/thomasbuchinger/timerec/internal/server/providers"
	"go.uber.org/zap"
)

func NewTestServer(mem *providers.FileOrMemoryProvider) server.TimerecServer {
	logger, _ := zap.NewDevelopment()

	return server.TimerecServer{
		Logger:        *logger.Sugar(),
		StateProvider: mem,
		TimeProvider:  mem,
		ChatProvider:  mem,
	}
}

// // Activity cannot be started, when another one is already active
// func TestStartActivityIfAnotherActivityIsActive(t *testing.T) {
// 	mem := providers.NewMemoryProvider()
// 	mem.Data.Users[0] = api.User{
// 		Name:     "me",
// 		Inactive: false,
// 		Activity: api.Activity{
// 			ActivityName:    "exists",
// 			ActivityComment: "comment",
// 			ActivityStart:   time.Now(),
// 			ActivityTimer:   time.Now(),
// 		},
// 	}

// 	mgr := NewTestServer(mem)

// 	response, err := mgr.StartActivity(context.TODO(), server.StartActivityParams{
// 		UserName:       "me",
// 		ActivityName:   "new",
// 		StartString:    "0m",
// 		EstimateString: "0m",
// 	})

// 	if mem.Data.User["me"].Activity.ActivityName != "exists" {
// 		t.Logf("ActivityName updated, despite error. got %s, expected exists", mem.Data.Users[0].Activity.ActivityName)
// 		t.Fail()
// 	}
// 	if response.Success != false {
// 		t.Logf("Response.Success was %t, expected false", response.Success)
// 		t.Fail()
// 	}
// 	if err == nil {
// 		t.Logf("Expected error, got nil: %s", err.Error())
// 		t.Fail()
// 	}
// }

// func TestStartActivityWorks(t *testing.T) {
// 	mem := providers.NewMemoryProvider()
// 	mgr := NewTestServer(mem)

// 	mgr.StateProvider.CreateUser(api.NewDefaultUser("me"))
// 	response, _ := mgr.StartActivity(context.TODO(), server.StartActivityParams{
// 		UserName:       "me",
// 		ActivityName:   "new",
// 		StartString:    "0m",
// 		EstimateString: "0m",
// 	})
// 	if mem.Data.User["me"].Activity.ActivityName != "new" {
// 		t.Logf("ActivityName not updated. got %s", mem.Data.User["me"].Activity.ActivityName)
// 		t.Fail()
// 	}
// 	if response.Activity.ActivityName != "new" {
// 		t.Logf("ActivityName not in response. got %s", response.Activity.ActivityName)
// 		t.Fail()
// 	}
// 	if response.Success != true {
// 		t.Logf("Response.Success was %t, expected true", response.Success)
// 		t.Fail()
// 	}
// }

// func TestExtendActivityWorks(t *testing.T) {
// 	mem := providers.NewMemoryProvider()
// 	dur, _ := time.ParseDuration("1h")
// 	mgr := NewTestServer(mem)

// 	mgr.StateProvider.CreateUser(api.NewDefaultUser("me"))
// 	mgr.StartActivity(context.TODO(), server.StartActivityParams{
// 		UserName:         "me",
// 		ActivityName:     "new",
// 		StartDuration:    dur,
// 		EstimateDuration: dur,
// 	})
// 	ts1 := mem.Data.User["me"].Activity.ActivityTimer
// 	mgr.ExtendActivity(context.TODO(), server.ExtendActivityParams{
// 		UserName:         "me",
// 		EstimateDuration: dur + dur,
// 	})
// 	if ts1 == mem.Data.User["me"].Activity.ActivityTimer {
// 		t.Fatal("timestamp not updated")
// 	}
// }

// func TestFinishActivityWorks(t *testing.T) {
// 	mem := providers.NewMemoryProvider()
// 	dur, _ := time.ParseDuration("30m")
// 	mgr := NewTestServer(mem)

// 	mgr.StateProvider.CreateUser(api.NewDefaultUser("me"))
// 	mgr.CreateJobIfMissing(context.TODO(), server.SearchJobParams{
// 		Name:          "testwork",
// 		Owner:         "me",
// 		StartedAfter:  -24 * time.Hour,
// 		StartedBefore: time.Duration(0),
// 	})
// 	mgr.StartActivity(context.TODO(), server.StartActivityParams{
// 		UserName:         "me",
// 		ActivityName:     "test",
// 		StartDuration:    dur,
// 		EstimateDuration: dur,
// 	})

// 	act := mem.Data.User["me"].Activity
// 	if err := act.CheckActivityActive(); err != nil {
// 		t.Fatal("StartActivity did not work", act)
// 	}

// 	res, fin_err := mgr.FinishActivity(context.TODO(), server.FinishActivityParams{
// 		UserName:     "me",
// 		JobName:      "testwork",
// 		ActivityName: "test",
// 		EndDuration:  dur,
// 	})
// 	act = mem.Data.User["me"].Activity
// 	if fin_err != nil {
// 		t.Logf("Expecting no Error, got %v", fin_err)
// 		t.Fail()
// 	}
// 	if err := act.CheckNoActivityActive(); err != nil {
// 		t.Log("FinishActivity did not clear Activity from Profile")
// 		t.Fail()
// 	}
// 	if !res.Success {
// 		t.Logf("FinishActivity returned %t", res.Success)
// 		t.Fail()
// 	}
// 	work, ok := mem.Data.Jobs["testwork"]
// 	if !ok || len(work.Activities) != 1 {
// 		t.Log("Job seems to be empty", work)
// 		t.Fail()
// 	}
// }

// // Activities cannot be finish without a Job
// func TestFinishActivityFailsWithoutJob(t *testing.T) {
// 	mem := providers.NewMemoryProvider()
// 	dur, _ := time.ParseDuration("30m")
// 	mgr := NewTestServer(mem)

// 	mgr.StateProvider.CreateUser(api.NewDefaultUser("me"))
// 	mgr.StartActivity(context.TODO(), server.StartActivityParams{
// 		UserName:         "me",
// 		ActivityName:     "test",
// 		StartDuration:    dur,
// 		EstimateDuration: dur,
// 	})

// 	act := mem.Data.User["me"].Activity
// 	if err := act.CheckActivityActive(); err != nil {
// 		t.Fatal("StartActivity did not work", act)
// 	}

// 	res, err := mgr.FinishActivity(context.TODO(), server.FinishActivityParams{
// 		UserName:     "me",
// 		JobName:      "testwork",
// 		ActivityName: "test",
// 		EndDuration:  dur,
// 	})
// 	act = mem.Data.User["me"].Activity
// 	if err == nil {
// 		t.Log("expected no error, got nothing")
// 		t.Fail()
// 	}
// 	if res.Success {
// 		t.Fatalf("FinishActivity returned. got %t expected %t", res.Success, false)
// 	}
// 	if err := act.CheckActivityActive(); err != nil {
// 		t.Fatalf("Operation failed, but Activity still cleared. CheckActivityActive() returned %v", err)
// 	}
// }
