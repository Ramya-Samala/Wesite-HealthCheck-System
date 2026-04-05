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
