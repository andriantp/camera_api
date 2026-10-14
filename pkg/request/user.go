package request

type Login struct {
	UserName string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`

	SessionID  string `json:"-"`
	IPAddress  string `json:"-"`
	UserAgent  string `json:"-"`
	ClientType string `json:"-"`
}

type UserActor struct {
	UserID    string
	SessionID string
	Role      string

	IPAddress  string
	UserAgent  string
	ClientType string
}

type ChangePassword struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
}

type CreateUser struct {
	Role     string `json:"role" validate:"required,uuid4"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type UpdateProfile struct {
	Username *string `json:"username" validate:"omitempty,min=3,max=100"`
	Role     *string `json:"role" validate:"omitempty,uuid4"`
}

type ResetPassword struct {
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type UpdateUserStatus struct {
	Active *bool `json:"active" validate:"required"`
}
