package server

import (
	"context"
	"errors"
	"math"
	"path"
	"reflect"
	"runtime"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

type ReconcileResult struct {
	Ok, Requeue bool
	RetryAfter  time.Duration
	Error       error
}

func (mgr *TimerecServer) ReconcileForever(ctx context.Context) {
	defaultInterval, _ := time.ParseDuration("5m")

	for {
		result := mgr.ReconcileOnce(ctx)

		if result.Requeue && result.RetryAfter < defaultInterval {
			SleepWithContext(ctx, result.RetryAfter)
		} else {
			SleepWithContext(ctx, defaultInterval)
		}
	}

}

func (mgr *TimerecServer) ReconcileOnce(ctx context.Context) ReconcileResult {
	// userReconcilers run per User and get a User object in their context
	userReconcilers := []func(context.Context) ReconcileResult{
		mgr.reconcileTimer,
		mgr.reconcileBegin,
		// mgr.reconcileTest,
	}
	// globalReconilers do not depend on a User
	globalReconcilers := []func(context.Context) ReconcileResult{}

	// Start Reconcilers
	var runningReconcilers int = 0
	returns := make(chan ReconcileResult)
	userList, _ := mgr.StateProvider.ListUsers()
	for _, f := range userReconcilers {
		for _, user := range userList {

			newCtx := context.WithValue(ctx, "user", user)
			go mgr.runReconcile(newCtx, returns, f)
			runningReconcilers++
		}
	}
	for _, f := range globalReconcilers {

		go mgr.runReconcile(ctx, returns, f)
		runningReconcilers++
	}

	// Collect ReconcileResult as the reconcilers finish
	runResult := ReconcileResult{Requeue: false, RetryAfter: time.Duration(math.MaxInt64)}
	var funcResult ReconcileResult
	for i := 0; i < runningReconcilers; i++ {
		funcResult = <-returns

		runResult.Ok = runResult.Ok && funcResult.Ok
		runResult.Requeue = runResult.Requeue || funcResult.Requeue
		if funcResult.Requeue && funcResult.RetryAfter < runResult.RetryAfter {
			runResult.RetryAfter = funcResult.RetryAfter
		}
	}
	return runResult // Global ReconcileResult
}

func (mgr *TimerecServer) runReconcile(ctx context.Context, c chan ReconcileResult, reconcileFunc func(context.Context) ReconcileResult) {
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
			c <- ReconcileResult{Ok: false, Requeue: false}
			return
		default:
			// Log Errurs
			if !result.Ok && result.Error == nil {
				logger.Infof("function returned NOT Ok, but returned no error %s", funcName)
			}
			if result.Error != nil {
				logger.Infof("error in function '%s': %s", funcName, result.Error.Error())
			}

			c <- result
		}
	}
}

func SleepWithContext(ctx context.Context, delay time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

func (mgr *TimerecServer) reconcileTimer(ctx context.Context) ReconcileResult {
	user, ok := ctx.Value("user").(api.User)
	if !ok {
		return ReconcileResult{Error: errors.New("Unable to read user from Context")}
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

func (mgr *TimerecServer) reconcileBegin(ctx context.Context) ReconcileResult {
	user, ok := ctx.Value("user").(api.User)
	if !ok {
		return ReconcileResult{Error: errors.New("Unable to read user from Context")}
	}
	now := time.Now()
	weekdays := user.Settings.Weekdays

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
	alarm := time.Now().Local().Truncate(day).Add(user.Settings.MissedWorkAlarm)
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
