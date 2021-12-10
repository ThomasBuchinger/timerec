
# Create File-Backend

```yaml
# db.yaml
profile: {}
tasks: {}
records: []
```

# Usage

```bash

# Start Task
# Start a timer on TASK_NAME, started 30 minutes ago and create a reminder in 1 hour
./timerec start TASK_NAME --start -30m --est 1h

# Wait for the reminder to finish. (The Terminal creates a Notification, if a command completes in a non-active windows)
./timerec wait

# Finish Task
# Finish task TASK_NAME, using the prev saved start-time and set end-time to right now. This also saves the task to a permanent location
./timerec fin TASK_NAME --end 0s
```
