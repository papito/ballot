package response

import "github.com/papito/ballot/ballot/model"

type HealthResponse struct {
	Status string `json:"status"`
}

type WsVoteStarted struct {
	Event string `json:"event"`
}

type WsVoteFinished struct {
	Users []model.User `json:"users"`
	Event string `json:"event"`
}

type WsNewUser struct {
	model.User
	Event  string `json:"event"`
}

type WsUserVote struct {
	Event    string `json:"event"`
	UserId   string `json:"user_id"`
}

type WsSession struct {
	Event        string `json:"event"`
	SessionState int    `json:"session_state"`
	Users []model.User  `json:"users"`
}

const (
	UserAddedEvent = "USER_ADDED"
	UserVotedEVent = "USER_VOTED"
	VoteStartedEVent = "VOTING"
	VoteFinishedEvent = "VOTE_FINISHED"
)

