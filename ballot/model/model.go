package model

type Session struct {
	// FIXME: make this "session_id"
	SessionId string `json:"id"`
}

type User struct {
	UserId   string `json:"id"`
	Name     string `json:"name"`
	Estimate int    `json:"estimate"`
}

type Vote struct {
	SessionId string `json:"session_id"`
	UserId 	  string `json:"user_id"`
	Estimate  int    `json:"estimate"`
}

type ValidationError struct {
	Field    string `json:"field"`
	ErrorStr string `json:"error"`
}

func (e ValidationError) Error() string {
	return e.ErrorStr
}

const (
	NotVoting = iota
	Voting
	FinishedVoting
)

const NoEstimate = -1