package model

type Session struct {
	// FIXME: make this "session_id"
	SessionId string `json:"id"`
}

type User struct {
	UserId   string `json:"id"`
	Name     string `json:"name"`
	Estimate int    `json:"estimate"`
	Voted    bool    `json:"voted"`
}

type PendingVote struct {
	SessionId string `json:"session_id"`
	UserId 	  string `json:"user_id"`
	// Pending vote has no estimate since we are hiding it while the vote is going on
}

type VoteResults struct {
	Votes[] User `json:"users"`
}

type ValidationError struct {
	Field    string `json:"field"`
	ErrorStr string `json:"error"`
}

// FIXME: move into its own custom error package
func (e ValidationError) Error() string {
	return e.ErrorStr
}

const (
	NotVoting = iota
	Voting
)

// FIXME: after the estimate is converted to string - this should just be empty string value
const NoEstimate = -1