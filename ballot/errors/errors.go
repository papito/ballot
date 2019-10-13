package errors


type ValidationError struct {
	Field    string `json:"field"`
	ErrorStr string `json:"error"`
}

func (e ValidationError) Error() string {
	return e.ErrorStr
}
