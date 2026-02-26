package model

// ApiResponse is the standard API response wrapper
type ApiResponse struct {
	Status   string      `json:"status"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Metadata *Metadata   `json:"metadata,omitempty"`
	Errors   string      `json:"errors,omitempty"`
}

// Metadata contains response metadata
type Metadata struct {
	LatencyMs int64  `json:"latency_ms"`
	Source    string `json:"source"`
}
