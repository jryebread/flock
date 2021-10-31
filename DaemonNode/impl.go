package DaemonNode

import (
	"encoding/json"
	"flock"
	"flock/Agent"
	"flock/IPFSPubSub"
	"flock/Node"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"strings"
	"time"
)



type DaemonInstance struct {
	NodeInstance *Node.NodeInstance
	FileLogger *logrus.Entry
	PubSub *IPFSPubSub.IPFSPubSubImpl
	Agent *Agent.AgentImpl
}


func CreateDaemonFromFile(file []byte) *DaemonInstance {

	data := &Node.NodeInstance{}

	_ = json.Unmarshal(file, data)
	fileLogger := flock.NewFileLogger(data.NodeDaemonPID)
	daemonLogger := fileLogger.Log.WithFields(logrus.Fields{
		"NodeDaemonPID": data.NodeDaemonPID,
	})
	daemonLogger.Info(fmt.Sprintf("Node UUID %s up!", data.Uuid))
	pubsub := IPFSPubSub.NewIPFSPubSub()

	//create agent
	agent := Agent.NewAgent(daemonLogger, pubsub.GetIpfsShell())
	return &DaemonInstance{
		NodeInstance: data,
		FileLogger: daemonLogger,
		PubSub: pubsub,
		Agent: agent,
	}
}

func (di *DaemonInstance) HandleAnnounceJoin() {
	for {
		di.FileLogger.Info("subbing to announce-join topic")
		subMsg, err := di.PubSub.SubscribeToTopic(flock.AnnounceJoinTopic)
		if err != nil {
			di.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("failure to sub to announce-join topic")
		}
		di.FileLogger.Info("subMsg received on announceJoinTopic: ", subMsg)
	}
}

func (di *DaemonInstance) HandleJobsReceive() {
	for {
		di.FileLogger.Info("subbing to job receive topic")
		subMsg, err := di.PubSub.SubscribeToTopic(flock.RecieveJobTopic)
		if err != nil {
			di.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("failure to sub to topic")
		}
		di.FileLogger.Info("subMsg received on jobs receive: ", subMsg)

		//TODO: Change to json job info
		s := strings.Split(subMsg, ":")
		node := s[0]
		job := s[1]
		if node == di.NodeInstance.NodeDaemonPID {
			di.FileLogger.Info("I recieved job: ", job)
			go di.enqueueJob(node, job)
		}
	}
}

//EnqueueJob is a goroutine that will enqueue the cid of the docker job to the replicated queue
// the smart scheduler will pull this job from the queue for execution later
//TODO: different job types beyond docker image tar files: binaries, straight up code files,
func (di *DaemonInstance) enqueueJob(nodeJobCreator string, jobCID string) {

	job := &flock.Job{
		NodeCreatorID: nodeJobCreator,
		CID:           jobCID,
		StartTime:     time.Now(),
		Status:        "started",
	}

	di.Agent.EnqueueJob(job)

	////TODO : Write the scheduler first and then come back to implement log sync

	fmt.Println("job ENQUEUED CID is: ", jobCID)

}

// RunAgent Runs the scheduler agent that executes tasks when the time is right, pops them off the queue, and stores them
// in history.json file.
func (di *DaemonInstance) RunAgent() {

	// on startup , fill up queue and completed_jobs w json history file
	// TODO: Read from history.json the completed tasks ->
	history := flock.AgentHistory{} //unmarshal file into this
	// and pass to agent so it can fill up its data structures from this file info

	go di.Agent.RunAgent(&history)
}

// HandleTaskStatusQuery query the agents completed_tasks hash set and work queue for the specific task
func (di *DaemonInstance) HandleTaskStatusQuery() {
	for {
		di.FileLogger.Info("subbing to TaskStatusQuery topic")
		subMsg, err := di.PubSub.SubscribeToTopic(flock.QueryForTaskStatusTopic)
		if err != nil {
			di.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("failure to sub to topic")
		}
		var taskInfo flock.TaskStatusQuery
		if err := json.Unmarshal([]byte(subMsg), &taskInfo); err != nil {
			di.FileLogger.Error("failure trying to unmarshall : ", err)
		}
		di.FileLogger.Info("subMsg received to query task status, checking agent: ", subMsg)
		//TODO: Check agent for completed_tasks and work queue for taskInfo
		// if found task, publish on Topic json of node task status:
		// (not_started, in_progress, done)
		// and results dir in IPFS (CID) and (started,end_job) timestamps if done or in progress
		di.PubSub.PublishToTopic(flock.TaskStatusResultTopic, subMsg)
	}
}

func (di *DaemonInstance) HandleJobsNew() {
	for {
		di.FileLogger.Info("subbing to NewJobs topic")
		subMsg, err := di.PubSub.SubscribeToTopic(flock.NewJobsTopic)
		if err != nil {
			di.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("failure to sub to topic")
		}
		di.FileLogger.Info("subMsg received on NewJobs, publishing on job score: ", subMsg)

		//TODO: Determine capability score based on hardware of this node
		// and the CPU usage of this node
		// di.Executor.GetCapabilityScore()
		min := 1
		max := 9
		score := rand.Intn(max - min) + min
		nodeName := di.NodeInstance.NodeDaemonPID
		res := nodeName + ":" + strconv.Itoa(score)
		di.FileLogger.Info("publishing score of ", strconv.Itoa(score))

		// publish node name and its score onto jobsScore topic
		di.PubSub.PublishToTopic(flock.JobNodeScoreTopic, res)
	}
}


