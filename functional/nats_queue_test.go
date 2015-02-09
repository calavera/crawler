package functional

import (
	"testing"

	"github.com/calavera/crawler/context"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GnatsdTestSuite struct {
	suite.Suite
	conn queue.Connection
}

func (s *GnatsdTestSuite) TestPubSub() {
	jobUUID := queue.UUID()

	w := make(chan bool)
	processor := func(q queue.Connection, d db.Connection, m *queue.Message) {
		assert.Equal(s.T(), "http://example.com", m.URL)
		assert.Equal(s.T(), jobUUID, m.JobUUID)
		assert.Equal(s.T(), 0, m.Depth)
		w <- true
	}

	s.conn.Subscribe(processor)
	s.conn.Publish(jobUUID, "http://example.com", 0)
	<-w
}

func TestGnatsdSuite(t *testing.T) {
	if h, ok := context.ParseNatsNodes(); ok {
		d, _ := db.NewMapConn()
		s := &GnatsdTestSuite{
			conn: context.ConnectNatsQueue(h, d),
		}
		suite.Run(t, s)
	}
}
