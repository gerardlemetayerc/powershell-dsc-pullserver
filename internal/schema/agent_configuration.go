package schema

// AgentConfigurationBinding models a configuration assignment for a node.
// The default assignment has schedule_type set to "none".
type AgentConfigurationBinding struct {
	AgentID                string  `json:"agent_id,omitempty"`
	ConfigurationName      string  `json:"configuration_name"`
	ScheduleType           string  `json:"schedule_type"`
	ScheduledAt            *string `json:"scheduled_at,omitempty"`
	NextExecutionStartAt   *string `json:"next_execution_start_at,omitempty"`
	NextExecutionEndAt     *string `json:"next_execution_end_at,omitempty"`
	NextExecutionCompleted bool    `json:"next_execution_completed,omitempty"`
	RecurrenceMinutes      *int    `json:"recurrence_minutes,omitempty"`
	WindowMinutes          int     `json:"window_minutes"`
	ScheduledLastAppliedAt *string `json:"scheduled_last_applied_at,omitempty"`
	LastRequestedAt        *string `json:"last_requested_at,omitempty"`
	LastExecutionStatus    *string `json:"last_execution_status,omitempty"`
	LastExecutionState     *string `json:"last_execution_state,omitempty"`
	LastExecutionAt        *string `json:"last_execution_at,omitempty"`
	Enabled                bool    `json:"enabled"`
}
