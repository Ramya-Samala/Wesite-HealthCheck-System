package main

type HealthCheck struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status,omitempty"`
	Code     *int32 `json:"code,omitempty"` // pointer so we can distinguish 0 from "not checked yet"
	Checked  int64  `json:"checked,omitempty"`
	Duration string `json:"duration,omitempty"`
	Error    string `json:"error,omitempty"`
}

// PagedResponse with pagination info
type PagedResponse struct {
	Items []HealthCheck `json:"items"`
	Page  int           `json:"page"`
	Total int           `json:"total"`
	Size  int           `json:"size"`
}

// ErrBody is what we send back when something goes wrong
type ErrBody struct {
	Message string `json:"message"`
}

// CreateReq is the expected body when creating a new check
type CreateReq struct {
	Endpoint string `json:"endpoint"`
}
