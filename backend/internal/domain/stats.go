package domain

type StatusCount struct {
	Status Status `json:"status"`
	Count  int64  `json:"count"`
}

type PriorityCount struct {
	Priority Priority `json:"priority"`
	Count    int64    `json:"count"`
}

type AgentResolvedCount struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	ResolvedCount int64  `json:"resolved_count"`
}

type AdminStats struct {
	TotalTickets       int64                `json:"total_tickets"`
	StatusBreakdown    []StatusCount        `json:"status_breakdown"`
	PriorityBreakdown  []PriorityCount      `json:"priority_breakdown"`
	SLABreachedCount   int64                `json:"sla_breached_count"`
	AvgResolutionHours float64              `json:"avg_resolution_hours"`
	TopAgents          []AgentResolvedCount `json:"top_agents"`
}
