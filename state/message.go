package state

import (
	"encoding/json"
	"time"
)

/*
type JobRequest struct {
	JobID       string
	RequestorID AgentID
	JobType     string
	JobPayload  json.RawMessage
}
*/
type JobResult struct {
	ResultID ResultID
	Result   string
	Arrived  time.Time
}

type Frame struct {
	SenderName    string
	SenderID      AgentID
	Timestamp     time.Time
	JobRequests   []*Job
	JobResults    []*JobResult
	MyPendingJobs []JobLabel
	MyDoneJobs    []JobLabel
	PubKey        string
	Signature     string
}

type JobLabel struct {
	ID   string
	Type JobType
}

func (f *Frame) FormToSign() ([]byte, error) {
	f.PubKey = ""
	f.Signature = ""
	return json.Marshal(f)
}

type AgentID string
