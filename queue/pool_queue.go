package queue

import "github.com/calavera/crawler/db"

// PoolConn implements queue.Connection using a channel as a backend.
// This interface is only suitable for testing.
// It offers no guarantees about the elements pushed and pulled from the queue.
type PoolConn struct {
	db db.Connection
	q  chan *Message
}

// NewPoolConn initializes the channel connection
func NewPoolConn(d db.Connection) Connection {
	return &PoolConn{
		db: d,
		q:  make(chan *Message),
	}
}

// Publish sends messages to the channel for a specific job
func (p *PoolConn) Publish(jobUUID, url string, depth uint) error {
	p.q <- NewMessage(jobUUID, url, depth)
	return nil
}

// Subscribe receives messages from the channel to process them
func (p *PoolConn) Subscribe(processor Processor) {
	go func() {
		for {
			select {
			case msg := <-p.q:
				go processor(p, p.db, msg)
			}
		}
	}()
}
