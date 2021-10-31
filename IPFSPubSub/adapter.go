package IPFSPubSub

type IPFSPubSub interface {
	PublishToTopic(topic string)
	SubscribeToTopic(topic string)
}