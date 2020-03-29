package brokers



type MessageBroker interface {
	Listen() chan *Message
}
