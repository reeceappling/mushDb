package checkers

type StartupStatus string

const (
	StartupStatusStarting StartupStatus = "starting"
	StartupStatusStarted  StartupStatus = "started"
)

type Started bool // True is started
const (
	StatusStarting Started = false
	StatusStarted  Started = true
)

func StartupStatusString(started Started) StartupStatus {
	if started == StatusStarted {
		return StartupStatusStarted
	}
	return StartupStatusStarting
}
