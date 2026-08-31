package passcli

import "context"

// AgentState is the daemon's lifecycle state.
type AgentState string

// Daemon states. AgentDegraded means the process is alive but its socket is
// missing; AgentUnknown means the status output could not be classified.
const (
	AgentRunning  AgentState = "running"
	AgentStopped  AgentState = "stopped"
	AgentDegraded AgentState = "degraded"
	AgentUnknown  AgentState = "unknown"
)

// AgentStatus is the parsed result of `ssh-agent daemon status`.
type AgentStatus struct {
	State AgentState
	// Detail is the full Status: value, including the parenthetical
	// explanation upstream adds for the degraded and stale-socket cases.
	Detail string
	// PID and Socket are empty when no PID file exists.
	PID    string
	Socket string
}

// AgentStatus reports the SSH agent daemon's state.
//
// The command has no JSON output and must be line-parsed, and it exits 0
// whether or not the daemon is running — so the state comes from stdout and
// never from the exit code. See docs/schema.md §7.
func (c *Client) AgentStatus(ctx context.Context) (AgentStatus, error) {
	stdout, err := c.call(ctx, "ssh-agent daemon status", nil, "ssh-agent", "daemon", "status")
	if err != nil {
		return AgentStatus{State: AgentUnknown}, err
	}

	return parseAgentStatus(stdout), nil
}

// StartAgent starts the SSH agent daemon in the background.
func (c *Client) StartAgent(ctx context.Context) error {
	_, err := c.call(ctx, "ssh-agent daemon start", nil, "ssh-agent", "daemon", "start")

	return err
}

// StopAgent stops the SSH agent daemon.
func (c *Client) StopAgent(ctx context.Context) error {
	_, err := c.call(ctx, "ssh-agent daemon stop", nil, "ssh-agent", "daemon", "stop")

	return err
}
