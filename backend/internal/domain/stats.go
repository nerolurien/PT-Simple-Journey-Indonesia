package domain

type StatusCount struct {
	Status Status `json:"status"`
	Count  int64  `json:"count"`
}

type PriorityCount struct {
	Priority Priority `json:"priority"`
	Count    int64    `json:"count"`
}

type AgentWorkload struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	Role            Role    `json:"role"`
	AssignedTotal   int64   `json:"assigned_total"`
	InProgressCount int64   `json:"in_progress_count"`
	ResolvedCount   int64   `json:"resolved_count"`
	CompletionRate  float64 `json:"completion_rate"`
}

type AgentResolvedCount = AgentWorkload

type AdminStats struct {
	TotalTickets       int64           `json:"total_tickets"`
	StatusBreakdown    []StatusCount   `json:"status_breakdown"`
	PriorityBreakdown  []PriorityCount `json:"priority_breakdown"`
	SLABreachedCount   int64           `json:"sla_breached_count"`
	AvgResolutionHours float64         `json:"avg_resolution_hours"`
	TopAgents          []AgentWorkload `json:"top_agents"`
}
