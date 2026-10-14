package response

import (
	"time"
)

type Profile struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type (
	ActivityUser struct {
		Total      int               `json:"total"`
		Returned   int               `json:"returned"`
		Activities ActivityUserRange `json:"activities"`
	}

	ActivityUserRange struct {
		Range Range      `json:"range"`
		Items []Activity `json:"items"`
	}

	Range struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	}

	Activity struct {
		UUID      string    `json:"uuid"`
		At        time.Time `json:"at"`
		Actor     *string   `json:"actor,omitempty"`
		IPAddress string    `json:"ip_address,omitempty"`
		UserAgent string    `json:"user_agent,omitempty"`
		Activity  string    `json:"activity"`
		Target    *string   `json:"target,omitempty"`
		Camera    *string   `json:"camera,omitempty"`
		Message   *string   `json:"message,omitempty"`
	}
)

type UserList struct {
	Total    int           `json:"total"`
	Returned int           `json:"returned"`
	Users    []UserSummary `json:"users"`
}

type (
	UserSummary struct {
		UUID         string        `json:"uuid"`
		Username     string        `json:"username"`
		Role         string        `json:"role"`
		LastLogin    *time.Time    `json:"last_login,omitempty"`
		LastSeen     *time.Time    `json:"last_seen,omitempty"`
		Active       bool          `json:"active"`
		CanUpdate    bool          `json:"can_update"`
		LastActivity *LastActivity `json:"last_activity,omitempty"`
	}
	LastActivity struct {
		At       *time.Time `json:"at"`
		Activity string     `json:"activity"`
		Message  *string    `json:"message,omitempty"`
	}
)
type RoleList struct {
	Total int64      `json:"total"`
	Roles []UUIDName `json:"users"`
}
