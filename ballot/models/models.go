package models

type Session struct {
	SessionId string `json:"id"`
}

// FIXME: export a factory that sets default values but not the struct
type User struct {
	UserId string `json:"id"`
	Name   string `json:"name"`
	Estimate  int `json:"estimate"`
}

type Vote struct {
	SessionId string `json:"session_id"`
	UserId 	  string `json:"user_id"`
	Estimate  int    `json:"estimate"`
}

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

const (
	NotVoting = iota
	Voting
	FinishedVoting
)

const NoEstimate = -1