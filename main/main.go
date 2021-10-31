package main

import (
	"flock"
	"flock/IPFSPubSub"
	"flock/Node"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)
//TODO: Change PID to use UUID I THINK

func FLOCK_READY() {
	fmt.Print(` FLOCK READY
    ('=   
    /_)  >')   >')   >')   >')   >')
   /||  <(_)> <(_)> <(_)> <(_)> <(_)>`)
	fmt.Println()
}
//go build -o flock main/main.go
func main() {
	args := os.Args
	if len(args) == 1 {
		log.Fatal("You should designate a cmd at the end such as `init` or `enqueue`")
	}
	cmd := args[1]
	pubsub := IPFSPubSub.NewIPFSPubSub()

	switch cmd {
	case "init":
		fmt.Println("Initializing Flock... ")
		// create a new node and store node info in a file in the home dir
		// node info includes uuid, capabilities (nodes cpu, gpu, ram, storage)
		fileSystem, closeFunc, err := flock.FileSystemUserStoreFromFile(flock.NodeStoreFileName)
		if err != nil {
			log.Fatal("Error opening node-info file")
		}
		newNode := Node.NewNodeInstance()
		pubsub.PublishToTopic(flock.AnnounceJoinTopic, newNode.Uuid)
		newNode.ReleaseNodeDaemon()
		newNode.WriteNodeToFile(fileSystem, closeFunc)
		FLOCK_READY()
		return
	case "add":
		if len(args) < 3 {
			log.Fatal("Error: user should designate task for add cmd")
		}
		//TODO: cmdline client should store the task info on the nodes daemon
		// this way if the sending of a task to another node fails,
		// this nodes daemon will still have the task and task will not be dropped
		job := args[2]
		// the client will subscribe and wait for node with highest cap score
		nodeForJob := make(chan string)
		go handleRecieveNodeScores(pubsub, nodeForJob)
		pubsub.PublishToTopic(flock.NewJobsTopic, args[2])
		node := <-nodeForJob // block for goroutine to finish and return node daemon id
		fmt.Println("Publishing to this node on jobRecieve topic: ", node)
		pubsub.PublishToTopic(flock.RecieveJobTopic, node + ":" + job)

	case "clear":
		fmt.Println("Terminating daemon..")
		//terminate all daemons
		pubsub.PublishToTopic(flock.TerminateTopic, "")

		// Clear the node_info file in home dir
		// Which will allow the user to do "init" cmd again
		// init will block while the file is already created

	}
}

type NodeForElection struct {
	nodePID string
	nodeScore int
}

//TODO: this goroutine will just hang, should find out how to clean it up,
// not too bad since this main program will exit anyway soon
func readFromTopic(quitCh chan bool, pubsub *IPFSPubSub.IPFSPubSubImpl, nodeScores *[]string) {
	for {
		select {
		case <- quitCh:
			fmt.Println("QUITTING READING FROM SUB GOROUTINE")
			return
		default:
			subMsg, err := pubsub.SubscribeToTopic(flock.JobNodeScoreTopic)
			if err != nil {
				log.Fatalf("failure trying to sub to topic for score postings %s", err.Error())
			}
			fmt.Println(subMsg)
			*nodeScores = append(*nodeScores, subMsg)
		}

	}
}

func handleRecieveNodeScores(pubsub *IPFSPubSub.IPFSPubSubImpl, nodeForJob chan string) {
	// find node with best score for the task, gather node and scores for a few seconds
	nodeScores := make([]string, 0)
	quitCh := make(chan bool)

	go readFromTopic(quitCh, pubsub, &nodeScores)
	fmt.Println("waiting on job score goroutine.. 2 seconds..")
	time.Sleep(2 * time.Second)

	nodeMax := &NodeForElection{
		nodeScore: -1,
	}
	//find node with highest score and publish them the job
	for _, msg := range nodeScores {
		fmt.Println("msg: ", msg)
		s := strings.Split(msg, ":")
		nodePID := s[0]
		nodeScore, _ := strconv.Atoi(s[1])

		n := &NodeForElection{nodePID, nodeScore}
		if n.nodeScore > nodeMax.nodeScore {
			//found new max node
			nodeMax = n
		}
	}
	fmt.Printf("Node %s wins the new job with a score of %v!\n", nodeMax.nodePID, nodeMax.nodeScore)
	nodeForJob <- nodeMax.nodePID
}