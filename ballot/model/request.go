package model

type CreateUserRequest struct {
	UserName string `json:"name"`
	SessionId string `json:"user_id"`
}

type StartVoteRequest struct {
	SessionId string `json:"user_id"`
}