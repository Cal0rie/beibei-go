package common

type NewRequest struct {
	Model    string       `json:"model"`
	Messages []ReqMessage `json:"messages"`
}

type ReqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
