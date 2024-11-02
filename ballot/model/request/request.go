package request

type CreateUserRequest struct {
	UserName   string `json:"name"`
	SessionId  string `json:"session_id"`
	IsObserver int    `json:"is_observer"`
	IsAdmin    int    `json:"is_admin"`
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
	Estimate  string `json:"estimate"`
}
