# User Guide
Beware: work in progress. Compile at your own risk

### Setup
Compile the binary with `go build`. Once compiled create a "db.yaml" file. This is where timerec stores state between invocations.

```yaml
# db.yaml
profile: {}
tasks: {}
records: []
```
### Usage
As a user, there are mostly 2 concepts to understand. There is **one default activity**, that is used to track the currently active task. **Tasks** are whatever work you do on a given day (e.g. working for projects, meetings, appointments, ...). There can be more tasks, however they should all be done at the end of the day. Tasks are what ultimately written to the Backend.

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

# Developer Guide
Welcome to this massively overengineered project! Ultimately this Application should run as a Server and integrate with a couple of different Services

* **CLI**: The CLI is the first and most basic client for this service. The CLI interacts with the ClientAPI directly in go.
* **Client API**: Implements most of the logic, but does not store any state. Usually this would be part of the CLI, but it might be required for the ChatBot later. Communicates to ServerAPI via "REST" (not yet)
* **Server API**: Stores all state and drives the integration provider. Communicates the Backend-Services via REST
* **State Provider**: Store internal State (Tasks, User-Settings)
* **TimeService**: Where to actually store the Recordings
* **NotificationService**: Outgoing communication
