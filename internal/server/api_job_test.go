package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/thomasbuchinger/timerec/internal/server"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

func TestJobIfMissingWorks(t *testing.T) {
	mem := providers.NewMemoryProvider()
	mgr := NewTestServer(mem)

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
	if len(mem.Data.Jobs) != 1 {
		t.Fatalf("incorrect number of Jobs: got %d expected %d", len(mem.Data.Jobs), 1)
	}
}

func TestCreateJobIfMissingIsIdempotent(t *testing.T) {
	mem := providers.NewMemoryProvider()
	mgr := NewTestServer(mem)

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
	if len(mem.Data.Jobs) != 1 {
		t.Fatalf("incorrect number of Jobs: got %d expected %d", len(mem.Data.Jobs), 1)
	}
}
