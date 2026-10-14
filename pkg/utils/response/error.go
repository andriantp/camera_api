package response

type APIError struct {
	StatusCode int
	Message    string
	Cause      interface{}
}

func (e APIError) Error() string {
	return e.Message
}
