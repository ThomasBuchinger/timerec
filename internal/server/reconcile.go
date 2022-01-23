package server

import (
	"context"
	"path"
	"reflect"
	"runtime"
	"time"
)

type ReconcileResult struct {
	Ok, Requeue bool
	RetryAfter  time.Duration
	Error       error
}

func (mgr *TimerecServer) Reconcile() {
	reconcilers := []func(context.Context) ReconcileResult{
		mgr.reconcileTimer,
		mgr.reconcileBegin,
		// mgr.reconcileTest,
	}
	ctx := context.Background()
	returns := make(chan bool)
	for _, f := range reconcilers {
		// this should run as go-routine, but that swallows logging
		go mgr.runReconcile(ctx, returns, f)
	}
	for range reconcilers {
		<-returns
	}
}

func (mgr *TimerecServer) runReconcile(ctx context.Context, c chan bool, reconcileFunc func(context.Context) ReconcileResult) {
	logger := mgr.Logger.Named("Reconciler")

	// predefine a result variable to be reused.
	result := ReconcileResult{Ok: true, Requeue: false, RetryAfter: 0, Error: nil}
	fullFuncName := runtime.FuncForPC(reflect.ValueOf(reconcileFunc).Pointer()).Name()
	funcName := path.Base(fullFuncName)
	for {

		// run reconcile function
		result = reconcileFunc(ctx)
		select {
		// Stop Execution if Context expired
		case <-ctx.Done():
			c <- false
			return
		default:
			// Log Errurs
			if !result.Ok && result.Error == nil {
				logger.Infof("function returned NOT Ok, but returned no error %s", funcName)
			}
			if result.Error != nil {
				logger.Infof("error in function '%s': %s", funcName, result.Error.Error())
			}

			// Requeue if required
			if !result.Requeue {
				// Done. return from function
				logger.Infof("Reconciled %s", funcName)
				c <- true
				return
			} else {
				// log, wait and continue infinite loop
				logger.Infof("requeueing %s", funcName)
				SleepWithContext(ctx, result.RetryAfter)
				continue
			}
		}
	}
}

func SleepWithContext(ctx context.Context, delay time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

func (mgr *TimerecServer) reconcileTimer(_ctx context.Context) ReconcileResult {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		return ReconcileResult{Error: err}
	}
	if user.Activity.ActivityTimer.IsZero() {
		return ReconcileResult{Ok: true} // No Timer Set, Nothing to do
	}
	timer := user.Activity.ActivityTimer
	if timer.After(time.Now()) {
		return ReconcileResult{Ok: true, Requeue: true, RetryAfter: time.Until(timer)} // Timer is in the furture. Waiting...
	}

	// Timer expired
	event := MakeEvent("TIMER_EXPIRED", "Estimated time expired", "activity@"+user.Activity.ActivityName, "me")
	err2 := mgr.ChatProvider.NotifyUser(event)
	if err2 != nil {
		return ReconcileResult{Error: err2}
	}
	return ReconcileResult{Ok: true}
}

func (mgr *TimerecServer) reconcileBegin(_ctx context.Context) ReconcileResult {
	now := time.Now()
	weekdays := mgr.Settings.Settings.Weekdays

	// Check for Weekdays
	isWeekday := false
	for _, day := range weekdays {
		if now.Weekday().String() == day {
			isWeekday = true
			break
		}
	}
	if !isWeekday {
		return ReconcileResult{Ok: true}
	}

	day, _ := time.ParseDuration("1d")
	alarm := time.Now().Local().Truncate(day).Add(mgr.Settings.Settings.MissedWorkAlarm)
	if now.Before(alarm) {
		return ReconcileResult{Ok: true, Requeue: true, RetryAfter: time.Until(alarm)}
	}

	event := MakeEvent("NO_ENTRY_ALARM", "No work logged today!", "activity@none", "me")
	err := mgr.ChatProvider.NotifyUser(event)
	if err != nil {
		return ReconcileResult{Error: err}
	}
	return ReconcileResult{Ok: true}
}

// func (mgr *TimerecServer) reconcileTest(_ context.Context) ReconcileResult {
// 	return ReconcileResult{Requeue: true}
// }
