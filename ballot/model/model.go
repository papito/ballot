package model

type Session struct {
	// FIXME: make this "session_id"
	SessionId string `json:"id"`
}

type User struct {
	UserId   string `json:"id"`
	Name     string `json:"name"`
	Estimate string `json:"estimate"`
	Voted    bool   `json:"voted"`
}

type PendingVote struct {
	SessionId string `json:"session_id"`
	UserId 	  string `json:"user_id"`
	// Pending vote has no estimate since we are hiding it while the vote is going on
}

type VoteResults struct {
	Votes[] User `json:"users"`
}

const (
	NotVoting = iota
	Voting
)

const NoEstimate = ""