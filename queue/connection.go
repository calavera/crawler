package queue

import "github.com/calavera/crawler/db"

// Processor defines a function interface to process messages.
type Processor func(Connection, db.Connection, *Message)

// Connection is an interface that defines how messages are published are received from a queue.
type Connection interface {
	// Publish pushes new messages to the queue.
	Publish(string, string, uint) error
	// Subscribe pulls messages from the queue and processes them using the processor function.
	Subscribe(Processor)
}
