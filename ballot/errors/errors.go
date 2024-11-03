package errors

type ValidationError struct {
	Field    string `json:"field"`
	ErrorStr string `json:"error"`
}

func (e ValidationError) Error() string {
	return e.ErrorStr
}

type CriticalError struct {
	Message string `json:"message"`
}

func (e CriticalError) Error() string {
	return e.Message
}
