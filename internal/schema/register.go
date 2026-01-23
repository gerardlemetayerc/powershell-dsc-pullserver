package schema

type RegisterRequest struct {
	NodeName string `json:"NodeName"`
}

type RegisterResponse struct {
	AgentId string `json:"AgentId"`
}
