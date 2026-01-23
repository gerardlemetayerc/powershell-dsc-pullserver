package schema

type Property struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

type NodeProperty struct {
	NodeName   string `json:"node_id"`
	PropertyID int    `json:"property_id"`
	Value      string `json:"value"`
}
