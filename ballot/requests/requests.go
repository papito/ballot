package requests

type CreateUserRequest struct {
	UserName string `json:"name"`
	SessionId string `json:"user_id"`
}
