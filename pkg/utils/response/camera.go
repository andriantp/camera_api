package response

import "time"

type ProtocolList struct {
	Total     int64      `json:"total"`
	Protocols []UUIDName `json:"protocols"`
}

type CameraList struct {
	Total   int64           `json:"total"`
	Cameras []CameraSummary `json:"cameras"`
}

type CameraSummary struct {
	UUID     string `json:"uuid"`
	SN       string `json:"sn"`
	Protocol string `json:"protocol"`
	Active   bool   `json:"active"`
}

type CameraDetail struct {
	UUID       string      `json:"uuid"`
	SN         string      `json:"sn"`
	Protocol   UUIDName    `json:"protocol"`
	Profile    *string     `json:"profile,omitempty"`
	Channel    *int        `json:"channel,omitempty"`
	Content    *Content    `json:"content,omitempty"`
	Active     bool        `json:"active"`
	Credential *Credential `json:"credential,omitempty"`
	Direction  *Direction  `json:"direction,omitempty"`
}

type Content struct {
	SearchID *int `json:"searchID"`
	TrackID  *int `json:"trackID"`
}

type Credential struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type (
	Direction struct {
		UUID  string   `json:"uuid"`
		Name  string   `json:"name"`
		Left  *PanTilt `json:"left,omitempty"`
		Right *PanTilt `json:"right,omitempty"`
		Up    *PanTilt `json:"up,omitempty"`
		Down  *PanTilt `json:"down,omitempty"`
	}

	DirectionDetail struct {
		UUID      string         `json:"uuid"`
		Name      string         `json:"name"`
		CreatedBy string         `json:"created_by"`
		CreatedAt time.Time      `json:"created_at"`
		UpdatedAt time.Time      `json:"updated_at"`
		Left      *PanTiltDetail `json:"left,omitempty"`
		Right     *PanTiltDetail `json:"right,omitempty"`
		Up        *PanTiltDetail `json:"up,omitempty"`
		Down      *PanTiltDetail `json:"down,omitempty"`
	}

	PanTilt struct {
		UUID     string  `json:"uuid"`
		Name     string  `json:"name"`
		Pan      float64 `json:"pan"`
		Tilt     float64 `json:"tilt"`
		Duration string  `json:"duration"`
	}
)
type (
	PanTiltSummary struct {
		UUID     string  `json:"uuid"`
		Name     string  `json:"name"`
		Pan      float64 `json:"pan"`
		Tilt     float64 `json:"tilt"`
		Duration string  `json:"duration"`
	}

	PanTiltList struct {
		Total    int64            `json:"total"`
		PanTilts []PanTiltSummary `json:"pan_tilts"`
	}

	PanTiltDetail struct {
		UUID      string    `json:"uuid"`
		Name      string    `json:"name"`
		Pan       float64   `json:"pan"`
		Tilt      float64   `json:"tilt"`
		Duration  string    `json:"duration"`
		CreatedBy string    `json:"created_by"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type (
	DirectionSummary struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}

	DirectionList struct {
		Total      int64              `json:"total"`
		Directions []DirectionSummary `json:"directions"`
	}
)

type Download struct {
	Filename    string    `json:"filename"`
	Size        float64   `json:"size"`
	ContentType string    `json:"contentType"`
	Expires     time.Time `json:"expires"`
	URL         string    `json:"url"`
}
