package response

type Success struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Failed struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Cause   interface{} `json:"cause,omitempty"`
}

type SuccessMeta struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Meta    Meta        `json:"meta,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Meta struct {
	TotalRecords int `json:"total_records"`
	Limit        int `json:"limit"`
	Offset       int `json:"offset"`
	Returned     int `json:"returned"`
}

func SuccessResponse(message string, data interface{}) Success {
	return Success{
		Status:  true,
		Message: message,
		Data:    data,
	}
}

func FailedResponse(message string, cause interface{}) Failed {
	return Failed{
		Status:  false,
		Message: message,
		Cause:   cause,
	}
}

func PaginationResponse(message string,meta Meta,data interface{}) SuccessMeta {
	return SuccessMeta{
		Status:  true,
		Message: message,
		Meta:    meta,
		Data:    data,
	}
}
