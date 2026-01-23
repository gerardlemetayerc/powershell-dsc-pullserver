package schema

type DscActionResponse struct {
	NodeStatus string           `json:"NodeStatus"`
	Details    []DscActionDetail `json:"Details"`
}
