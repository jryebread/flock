package Agent

import (
	"flock"
	"fmt"
	"github.com/enriquebris/goconcurrentqueue"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)


type AgentImpl struct {
	completedJobs *flock.Set
	workQ goconcurrentqueue.Queue
	FileLogger *logrus.Entry
	IpfsShell *shell.Shell
}


func NewAgent(fileLogger *logrus.Entry, shell *shell.Shell) *AgentImpl {
	return &AgentImpl{
		completedJobs: flock.NewSet(),
		workQ: goconcurrentqueue.NewFIFO(),
		FileLogger: fileLogger,
		IpfsShell : shell,
	}
}

func (i *AgentImpl) EnqueueJob(job *flock.Job) {
	//TODO: this enqueue will fail if the queue is locked, implement retry
	i.workQ.Enqueue(job)
}

var IpfsOutputDirDockerTmp = ""
var IpfsOutputFileDir = ""

// RunAgent goroutine Periodically sleeps, on wake checks nodes compute capacity, so it can determine which tasks to execute
// on startup, loads data structures with history
func (i *AgentImpl) RunAgent(history *flock.AgentHistory) {
	//TODO: Fill up q and set with history data
	dirname, err := os.UserHomeDir()
	if err != nil {
		i.FileLogger.Error("problem reading user home dir")
	}
	i.FileLogger.Info("Setting up agent...")
	IpfsOutputDirDockerTmp = dirname + "/tmp_docker_output/"
	IpfsOutputFileDir = dirname + "/ipfs_file_output/"
	_ = os.Mkdir(IpfsOutputDirDockerTmp, 0777) //err will say dir already exists
	_ = os.Mkdir(IpfsOutputFileDir, 0777) //err will say dir already exists
	for {
		time.Sleep(5 * time.Second) //sleep periodically
		i.GetNodeCapacityScore()

		//retrieve next job from queue
		jobPop, err := i.workQ.Dequeue()
		if err != nil {
			//i.FileLogger.WithFields(logrus.Fields{"error": err}).Error("failure: queue was either locked or empty")
			continue
		}
		i.FileLogger.Info("Got past, ", jobPop)
		job := jobPop.(*flock.Job)
		//retrieve docker image file with CID FROM IPFS
		//write file to local
		i.FileLogger.Info(job.CID)

		res := i.IpfsShell.Get(job.CID, IpfsOutputDirDockerTmp)
		i.FileLogger.WithFields(logrus.Fields{"result from ipfs_get": fmt.Sprintf("failed trying to get CID %v from ipfs", res) })
		if err != nil {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error(fmt.Sprintf("failed trying to get CID %v from ipfs", res))
			continue
		}

		files, err := ioutil.ReadDir(IpfsOutputDirDockerTmp)
		if err != nil {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("problem reading the ipfs output dir")
			continue
		}
		if len(files) != 1 {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("problem reading tpm out/ dir, missing single docker image file")
			continue
		}
		dockerFile := files[0]

		//docker load
		out, err:= loadDockerImage(dockerFile.Name())
		if err != nil {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("problem reading tmp /out/ dir")
			continue
		}
		imageName := parseLoadDockerOut(out)
		i.FileLogger.Info("Docker image loaded: ", imageName)


		out, err = executeDocker(imageName)
		//TODO: IDEA: should we add possibility for non docker, and execute any normal code file in a lambda docker?
		if err != nil {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("problem executing docker run")
			continue
		}
		i.FileLogger.WithFields(logrus.Fields{"docker_output": out}).Info("docker run output captured")

		//run docker run on local image
		//execute the docker image as a container that should exit and return out
		CID, err := i.CreateIPFSFileForOutput(job, out)
		if err != nil {
			i.FileLogger.WithFields(logrus.Fields{"error": err}).
				Error("problem writing output file to ipfs")
			continue
		}
		job.JobOutputFilesCID = CID
		//TODO: Update job object with other info before adding it to the map of completed_tasks (jobCID -> jobObj)

		////Clear the temp docker output dir (only one docker image should be in there)
		//dir, err := ioutil.ReadDir(IpfsOutputDirDockerTmp)
		//if err != nil {
		//	i.FileLogger.WithFields(logrus.Fields{"error": err}).
		//		Error("problem removing home dir ipfs_out temp ")
		//	continue
		//}
		//for _, d := range dir {
		//	os.RemoveAll(path.Join([]string{IpfsOutputDirDockerTmp, d.Name()}...))
		//}


	}

}

func parseLoadDockerOut(out string) string {
	return out[14:] //get image name from loadDocker cmd output
}

func loadDockerImage(filename string) (string, error){
	cmd := "docker"
	args := []string{
		"load", "--input", IpfsOutputDirDockerTmp + filename}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func executeDocker(imageName string) (string, error) {
	cmd := "docker"
	args := []string{
		"run", SpaceMap(imageName)} // RUN ANY DOCKER IMAGE THAT WAS LOADED
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

//removes spaces from string
func SpaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func executeDockerImagesCmd() string {
	cmd := "docker"
	args := []string{
		"images"}
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Println("problem running docker images cmd")
		log.Fatal(err)
	}
	return string(out)
}

//TODO: Implement node capacity based on cpu usage
func (i *AgentImpl) GetNodeCapacityScore() int {
	return 5
}

func (i *AgentImpl) CreateIPFSFileForOutput(job *flock.Job, out string) (string, error) {
	// write out to ipfs file, save cid in job object
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("./tmp_out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		i.FileLogger.WithFields(logrus.Fields{"error": err}).
			Error("problem reading tmp /out/ dir")
		return "",  err
	}

	_, err = f.Write([]byte(out))
	if err != nil {
		i.FileLogger.WithFields(logrus.Fields{"error": err}).
			Error("problem reading tmp /out/ dir")
		return "", err
	}

	CID, err := i.IpfsShell.Add(f)
	f.Close()
	return CID, nil
}
