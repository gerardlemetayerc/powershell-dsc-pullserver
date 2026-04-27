package schema

// AgentConfigurationBinding models a configuration assignment for a node.
// The default assignment has schedule_type set to "none".
type AgentConfigurationBinding struct {
	AgentID                string  `json:"agent_id,omitempty"`
	ConfigurationName      string  `json:"configuration_name"`
	ScheduleType           string  `json:"schedule_type"`
	ScheduledAt            *string `json:"scheduled_at,omitempty"`
	RecurrenceMinutes      *int    `json:"recurrence_minutes,omitempty"`
	WindowMinutes          int     `json:"window_minutes"`
	ScheduledLastAppliedAt *string `json:"scheduled_last_applied_at,omitempty"`
	Enabled                bool    `json:"enabled"`
}
