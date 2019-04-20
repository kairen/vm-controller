package types

type Server struct {
	ID       int32  `json:"id"`
	UUID     string `json:"uuid"`
	Name     string `json:"name" binding:"required"`
	CPU      int    `json:"cpu,omitempty" binding:"required"`
	Memory   int    `json:"memory,omitempty" binding:"required"`
	DiskSize int    `json:"diskSize,omitempty" binding:"required"`
}

type ServerState string

const (
	ServerNoState          ServerState = "No"
	ServerRunningState     ServerState = "Running"
	ServerBlockedState     ServerState = "Blocked"
	ServerStopState        ServerState = "Stop"
	ServerPauseState       ServerState = "Pause"
	ServerShutdownState    ServerState = "Shutdown"
	ServerShutoffState     ServerState = "Shutoff"
	ServerCrashedState     ServerState = "Crashed"
	ServerPmsuspendedState ServerState = "Pmsuspended"
)

type ServerStatus struct {
	CPUUtilization int         `json:"cpuUtilization" binding:"required"`
	State          ServerState `json:"state" binding:"required"`
}
