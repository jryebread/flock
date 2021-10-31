package main

import (
	"flock"
	"flock/DaemonNode"
	"fmt"
	"io/ioutil"
)

const fileName = "flock-node-info.json"

func main() {
	// this is the main loop for the node worker daemon
	fmt.Println("HEHE")
	file, _ := ioutil.ReadFile(fileName)
	daemon := DaemonNode.CreateDaemonFromFile(file)

	go daemon.HandleAnnounceJoin()
	go daemon.HandleJobsNew()
	go daemon.HandleJobsReceive()
	go daemon.HandleTaskStatusQuery()

	daemon.RunAgent()


	// terminate this daemon on recieving, otherwise hold forever
	// so the other daemon topic handler goroutines can run
	_, _ = daemon.PubSub.SubscribeToTopic(flock.TerminateTopic)
	daemon.FileLogger.Info("got terminate cmd, terminating daemon.. ", )
}