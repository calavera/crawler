package queue

import (
	"testing"

	"github.com/calavera/crawler/db"
	"github.com/stretchr/testify/assert"
)

func TestPoolConn(t *testing.T) {
	d, _ := db.NewMapConn()
	q := NewPoolConn(d)

	done := make(chan bool)
	processor := func(q Connection, d db.Connection, m *Message) {
		d.Save(m.JobUUID, m.URL)
		done <- true
	}

	q.Subscribe(processor)

	q.Publish("test", "http://example.com", 0)
	<-done

	r, _ := d.Results("test")
	assert.Equal(t, 1, len(r))
}
