package redpanda

type UserRegisteredEvent struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Code    string `json:"code"`
}

type VerificationCodeUpdatedEvent struct {
	Email string `json:"user_id"`
	Code  string `json:"code"`
}
