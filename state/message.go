package state

import "time"

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
	SenderName  string
	SenderID    AgentID
	JobRequests []*Job
	JobResults  []*JobResult
	PubKey      string
	Signature   string
}

type AgentID string
