package queue

import (
	"github.com/apcera/nats"
	"github.com/calavera/crawler/db"
)

const (
	crawlerTopic = "crawl-url"
	queueName    = "crawler-queue"
)

// NatsConn implements the queue.Connection interface using Gnatsd as a queue.
type NatsConn struct {
	db   db.Connection
	conn *nats.EncodedConn
	proc Processor
}

// NewNatsConn initializes the connection to Gnatsd.
// It assumes that the client has been initialized by the application context.
func NewNatsConn(d db.Connection, conn *nats.EncodedConn) Connection {
	return &NatsConn{
		db:   d,
		conn: conn,
	}
}

// Publish enqueues new messages in the queue for a given job.
func (q *NatsConn) Publish(jobUUID, url string, depth uint) error {
	msg := NewMessage(jobUUID, url, depth)
	return q.conn.Publish(crawlerTopic, msg)
}

// Subscribe subscribes the job group to a specific topic to process messages.
func (q *NatsConn) Subscribe(processor Processor) {
	q.proc = processor
	q.conn.QueueSubscribe(crawlerTopic, queueName, q.processMessage)
}

func (q *NatsConn) processMessage(m *Message) {
	go q.proc(q, q.db, m)
}
