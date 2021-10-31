package Node

import (
	"flock"
	"flock/PeerList"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type NodeInstance struct {
	Uuid string `json:"uuid"`
	NodeDaemonPID string `json:"nodeDaemonPID"`
	PeerList *PeerList.PeerListImpl `json:"peerList"`
}

func (i *NodeInstance) WriteNodeToFile(fileSystem *flock.FileSystemUserStore, closeFunc func()) {
	err := fileSystem.DatabaseWriter.Encode(i)
	if err != nil {
		log.Fatal("failure trying to write node info to file")
	}
}

// creates a daemon sub-process that will continue to run as a worker
// returns its pid so that the flock main program can kill it on clear call from client
// how to keep track of daemons? One option is
func (i *NodeInstance) ReleaseNodeDaemon() {
	fmt.Println("starting flock daemon..")
	cmd := exec.Command("go", "run", "main/daemon.go")
	cmd.Stdout = os.Stdout
	err := cmd.Start() // run daemon without waiting on it
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cmd.Process.Pid)
	i.NodeDaemonPID = strconv.Itoa(cmd.Process.Pid)
}

func NewNodeInstance() *NodeInstance {
	uniqueIDForNewNode, _ := uuid.NewUUID()
	peerList := PeerList.InitPeerList()

	return &NodeInstance{
		Uuid: uniqueIDForNewNode.String(),
		PeerList: peerList,
	}
}