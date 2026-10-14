package request

type (
	CreateProtocol struct {
		Name string `json:"name" validate:"required,min=3,max=100"`
	}

	CreateCamera struct {
		SN       string `json:"sn" validate:"required,min=1,max=100"`
		Protocol string `json:"protocol" validate:"required,uuid4"`
		Host     string `json:"host" validate:"required"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
)

type (
	UpdateCameraSN struct {
		SN string `json:"sn" validate:"required,min=1,max=100"`
	}

	UpdateCameraStatus struct {
		Active *bool `json:"active" validate:"required"`
	}

	UpdateCameraCredential struct {
		Host     *string `json:"host" validate:"omitempty,max=255"`
		Username *string `json:"username" validate:"omitempty,max=100"`
		Password *string `json:"password" validate:"omitempty,max=100"`
	}

	UpdateCameraConfiguration struct {
		Target  string   `json:"target" validate:"required,oneof=profile channel content"`
		Profile *string  `json:"profile,omitempty"`
		Channel *int     `json:"channel,omitempty"`
		Content *Content `json:"content,omitempty"`
	}

	Content struct {
		SearchID *int `json:"searchID"`
		TrackID  *int `json:"trackID"`
	}

	UpdateCameraDirection struct {
		Direction string `json:"direction" validate:"required,uuid4"`
	}
)

type RelativeMove struct {
	Pan      *float64 `json:"pan"`
	Tilt     *float64 `json:"tilt"`
	Duration string   `json:"duration" validate:"required"`
}

type (
	CreatePanTilt struct {
		Name     string   `json:"name" validate:"required,min=1,max=100"`
		Pan      *float64 `json:"pan" validate:"required"`
		Tilt     *float64 `json:"tilt" validate:"required"`
		Duration string   `json:"duration" validate:"required"`
	}
	UpdatePanTilt struct {
		Name     *string  `json:"name" validate:"omitempty,min=1,max=100"`
		Pan      *float64 `json:"pan" validate:"omitempty"`
		Tilt     *float64 `json:"tilt" validate:"omitempty"`
		Duration *string  `json:"duration" validate:"omitempty"`
	}
)

type (
	CreateDirection struct {
		Name  string  `json:"name" validate:"required,min=1,max=100"`
		Left  *string `json:"left" validate:"omitempty,uuid4"`
		Right *string `json:"right" validate:"omitempty,uuid4"`
		Up    *string `json:"up" validate:"omitempty,uuid4"`
		Down  *string `json:"down" validate:"omitempty,uuid4"`
	}

	UpdateDirection struct {
		Name  *string `json:"name" validate:"omitempty,min=1,max=100"`
		Left  *string `json:"left" validate:"omitempty,uuid4"`
		Right *string `json:"right" validate:"omitempty,uuid4"`
		Up    *string `json:"up" validate:"omitempty,uuid4"`
		Down  *string `json:"down" validate:"omitempty,uuid4"`
	}
)

type AssignCameraRequest struct {
	CameraUUID string `json:"cameraUUID" validate:"required,uuid4"`
	UserUUID   string `json:"userUUID" validate:"required,uuid4"`
}
