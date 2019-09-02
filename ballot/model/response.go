package model

type HealthResponse struct {
	Status string `json:"status"`
}

type WsVoteStarted struct {
	Event string `json:"event"`
}

type WsUser struct {
	User
	Event  string `json:"event"`
}

type WsUserVote struct {
	Event string `json:"event"`
	UserId string `json:"user_id"`
	Estimate uint8 `json:"estimate"`
}

type WsSession struct {
	Event string       `json:"event"`
	SessionState int   `json:"session_state"`
	Users []User `json:"users"`
}

