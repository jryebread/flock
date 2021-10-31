package flock

import "time"

const (
	AnnounceJoinTopic = "announce-join"
	NewJobsTopic = "new-jobs"
	TerminateTopic = "terminate-node"
	JobNodeScoreTopic = "node-score"
	RecieveJobTopic = "recieve-job"
	QueryForTaskStatusTopic = "task-status"
	TaskStatusResultTopic = "task-status-result"

	NodeStoreFileName = "flock-node-info.json"


)

// We can use a struct to define a collection schema for the task
type Job struct {
	NodeCreatorID string `json:"node_creator_id"`
	CID string `json:"cid"`
	StartTime time.Time
	EndTime time.Time
	JobOutputFilesCID string
	Status string
}

type TaskStatusQuery struct {

}

type AgentHistory struct {

}