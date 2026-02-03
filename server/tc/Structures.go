package tc

import (
	"net/http"
	"time"
)

type Client struct {
	BaseURL      string
	HTTPClient   *http.Client
	PollInterval time.Duration
}

type response struct {
	Ack       bool   `json:"ack"`
	ErrorMsg  string `json:"error_msg"`
	ExeStatus string `json:"exe_status"`
}

type createRequest struct {
	ProcName  string `json:"proc_name"`
	Procedure string `json:"procedure"`
}

type validateRequest struct {
	ProcName  string `json:"proc_name"`
	ProcSrc   string `json:"proc_src"`
	Subsystem string `json:"subsystem"`
}

type loadRequest struct {
	Action       string `json:"action"`
	ProcName     string `json:"proc_name"`
	ProcSrc      string `json:"proc_src"`
	ProcMode     string `json:"proc_mode"`
	ProcPriority string `json:"proc_priority"`
}

type ProcedureResult struct {
	Success bool
	Status  Status
	Err     error
}

type Status string

const (
	StatusSuccess      Status = "success"
	StatusFailure      Status = "failure"
	StatusAborted      Status = "aborted"
	StatusSuspended    Status = "suspended"
	StatusNotAvailable Status = "not-available"
	StatusInProgress   Status = "in-progress"
	StatusQueued       Status = "queued"
	StatusUnknown      Status = "unknown"
)
