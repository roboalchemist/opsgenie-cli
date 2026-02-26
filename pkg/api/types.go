package api

import "fmt"

// APIResponse is the generic OpsGenie API response wrapper.
type APIResponse[T any] struct {
	Result    string  `json:"result,omitempty"`
	Took      float64 `json:"took,omitempty"`
	RequestID string  `json:"requestId,omitempty"`
	Data      T       `json:"data,omitempty"`
	Paging    *Paging `json:"paging,omitempty"`
}

// PagingInfo contains pagination metadata.
type Paging struct {
	First string `json:"first,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Last  string `json:"last,omitempty"`
}

// ErrorResponse represents an OpsGenie API error.
type ErrorResponse struct {
	Message string   `json:"message"`
	Code    int      `json:"code"`
	Errors  []string `json:"errors,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("OpsGenie API error %d: %s (details: %v)", e.Code, e.Message, e.Errors)
	}
	return fmt.Sprintf("OpsGenie API error %d: %s", e.Code, e.Message)
}

// RequestResult is the response from polling an async request.
type RequestResult struct {
	IsSuccess bool   `json:"isSuccess"`
	Status    string `json:"status"`
	AlertID   string `json:"alertId,omitempty"`
	Alias     string `json:"alias,omitempty"`
}

// AlertResponse represents a single alert.
type AlertResponse struct {
	ID          string            `json:"id"`
	TinyID      string            `json:"tinyId,omitempty"`
	Alias       string            `json:"alias,omitempty"`
	Message     string            `json:"message"`
	Status      string            `json:"status"`
	Acknowledged bool             `json:"acknowledged"`
	Snoozed     bool              `json:"snoozed,omitempty"`
	IsSeen      bool              `json:"isSeen,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Count       int               `json:"count,omitempty"`
	Source      string            `json:"source,omitempty"`
	Owner       string            `json:"owner,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Responders  []Responder       `json:"responders,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	UpdatedAt   string            `json:"updatedAt,omitempty"`
	ClosedAt    string            `json:"closedAt,omitempty"`
}

// Responder is a team or user assigned to an alert or incident.
type Responder struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type"` // "team" or "user"
}

// IncidentResponse represents a single incident.
type IncidentResponse struct {
	ID          string      `json:"id"`
	TinyID      string      `json:"tinyId,omitempty"`
	Message     string      `json:"message"`
	Status      string      `json:"status"`
	Tags        []string    `json:"tags,omitempty"`
	Owner       string      `json:"owner,omitempty"`
	Priority    string      `json:"priority,omitempty"`
	Responders  []Responder `json:"responders,omitempty"`
	Description string      `json:"description,omitempty"`
	CreatedAt   string      `json:"createdAt,omitempty"`
	UpdatedAt   string      `json:"updatedAt,omitempty"`
}

// TeamResponse represents a single team.
type TeamResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Members     []TeamMember `json:"members,omitempty"`
	Links       TeamLinks    `json:"links,omitempty"`
}

// TeamMember is a member of a team.
type TeamMember struct {
	User UserRef `json:"user"`
	Role string  `json:"role,omitempty"`
}

// UserRef is a reference to a user.
type UserRef struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

// TeamLinks contains hypermedia links for a team.
type TeamLinks struct {
	Web string `json:"web,omitempty"`
	API string `json:"api,omitempty"`
}

// ScheduleResponse represents a single schedule.
type ScheduleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	Enabled     bool   `json:"enabled"`
	OwnerTeam   *TeamRef `json:"ownerTeam,omitempty"`
}

// TeamRef is a reference to a team.
type TeamRef struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// UserResponse represents a single user.
type UserResponse struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	FullName  string   `json:"fullName,omitempty"`
	Role      UserRole `json:"role,omitempty"`
	Blocked   bool     `json:"blocked,omitempty"`
	Verified  bool     `json:"verified,omitempty"`
	CreatedAt string   `json:"createdAt,omitempty"`
}

// UserRole is the role assigned to a user.
type UserRole struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// HeartbeatResponse represents a single heartbeat monitor.
type HeartbeatResponse struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Interval      int    `json:"interval,omitempty"`
	IntervalUnit  string `json:"intervalUnit,omitempty"`
	Enabled       bool   `json:"enabled"`
	Expired       bool   `json:"expired,omitempty"`
	LastPingAt    string `json:"lastPingAt,omitempty"`
	AlertMessage  string `json:"alertMessage,omitempty"`
	AlertPriority string `json:"alertPriority,omitempty"`
	AlertTags     []string `json:"alertTags,omitempty"`
}

// EscalationResponse represents a single escalation policy.
type EscalationResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	OwnerTeam   *TeamRef            `json:"ownerTeam,omitempty"`
	Rules       []EscalationRule    `json:"rules,omitempty"`
}

// EscalationRule is one step in an escalation policy.
type EscalationRule struct {
	Condition  string    `json:"condition,omitempty"`
	NotifyType string    `json:"notifyType,omitempty"`
	Delay      DelayInfo `json:"delay,omitempty"`
	Recipient  Responder `json:"recipient,omitempty"`
}

// DelayInfo describes the time delay in an escalation rule.
type DelayInfo struct {
	TimeAmount int    `json:"timeAmount,omitempty"`
	TimeUnit   string `json:"timeUnit,omitempty"`
}

// IntegrationResponse represents a single integration.
type IntegrationResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	Enabled bool   `json:"enabled"`
}
