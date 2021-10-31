package IPFSPubSub

import (
	shell "github.com/ipfs/go-ipfs-api"
	"log"
)

type IPFSPubSubImpl struct {
	sh *shell.Shell
}

func (i *IPFSPubSubImpl) GetIpfsShell() *shell.Shell {
	return i.sh
}

func NewIPFSPubSub() *IPFSPubSubImpl{
	var sh = shell.NewShell("localhost:5001")
	return &IPFSPubSubImpl{
		sh: sh,
	}
}

func (i *IPFSPubSubImpl) PublishToTopic(topic string, val string) {
	err := i.sh.PubSubPublish(topic, val)
	if err != nil {
		log.Fatalf("Error trying to publish %v", err)
	}
}

func (i *IPFSPubSubImpl) SubscribeToTopic(topic string) (string, error) {
	sub, err := i.sh.PubSubSubscribe(topic)
	if err != nil {
		return "", err
	}
	subMsg, err := sub.Next()
	if err != nil {
		return "", err
	}
	return string(subMsg.Data), nil
}