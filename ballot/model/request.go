package model

type CreateUserRequest struct {
	UserName string `json:"name"`
	SessionId string `json:"session_id"`
}

type StartVoteRequest struct {
	SessionId string `json:"session_id"`
}

type CastVoteRequest struct {
	UserId string `json:"user_id"`
	SessionId string `json:"session_id"`
	Estimate uint8 `json:"estimate"`
}
