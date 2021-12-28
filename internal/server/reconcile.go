package server

import (
	"reflect"
	"runtime"
	"time"
)

type ReconcileResult struct {
	Ok, Requeue bool
	RetryAfter  float64
	Error       error
}

func (mgr *TimerecServer) Reconcile() {
	reconcilers := []func() ReconcileResult{
		mgr.reconcileTimer,
	}
	for _, f := range reconcilers {
		// this should run as go-routine, but that swallows logging
		mgr.runReconcile(f)
	}
}

func (mgr *TimerecServer) runReconcile(reconcileFunc func() ReconcileResult) {
	// start with Requue=true, to start the loop
	result := ReconcileResult{Ok: true, Requeue: true, RetryAfter: 0, Error: nil}
	funcName := runtime.FuncForPC(reflect.ValueOf(reconcileFunc).Pointer()).Name()
	for result.Requeue {
		result = reconcileFunc()

		if !result.Ok && result.Error == nil {
			mgr.logger.Printf("function returned NOT Ok, but returned no error %s \n", funcName)
		}
		if result.Error != nil {
			mgr.logger.Printf("error in function '%s': %s\n", funcName, result.Error.Error())
		}
		if result.Requeue {
			mgr.logger.Printf("requueing %s", funcName)
			time.Sleep(5 * time.Second)
		}
	}
	mgr.logger.Printf("Reconciled %s\n", funcName)
}

func (mgr TimerecServer) reconcileTimer() ReconcileResult {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		return ReconcileResult{Error: err}
	}
	if user.Activity.ActivityTimer.IsZero() {
		return ReconcileResult{Ok: true}
	}
	timer := user.Activity.ActivityTimer
	if timer.Before(time.Now()) {
		event := MakeEvent("TIMER_EXPIRED", "Estimated time expired", "activity@"+user.Activity.ActivityName, "me")
		err2 := mgr.ChatProvider.NotifyUser(event)
		if err2 != nil {
			return ReconcileResult{Error: err2}
		}
	}

	return ReconcileResult{Ok: true}
}
