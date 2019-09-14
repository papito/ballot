package request

type CreateUserRequest struct {
	UserName  string `json:"name"`
	SessionId string `json:"session_id"`
}

type StartVoteRequest struct {
	SessionId string `json:"session_id"`
}

type FinishVoteRequest struct {
	SessionId string `json:"session_id"`
}

type CastVoteRequest struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
	Estimate  int    `json:"estimate"`
}
