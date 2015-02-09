package functional

import (
	"testing"

	"github.com/calavera/crawler/context"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RiakTestSuite struct {
	suite.Suite
	conn    db.Connection
	jobUUID string
}

func (s *RiakTestSuite) SetupTest() {
	s.jobUUID = queue.UUID()

	err := s.conn.CreateJob(s.jobUUID)
	assert.NoError(s.T(), err)
}

func (s *RiakTestSuite) TestSimpleStatus() {
	i, err := s.conn.Status(s.jobUUID)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), i)
}

func (s *RiakTestSuite) TestViewPage() {
	_, err := s.conn.ViewPage(s.jobUUID, "http://example.com")
	assert.NoError(s.T(), err)

	i, err := s.conn.Status(s.jobUUID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "http://example.com", i.PageViews()[0].URL)
}

func (s *RiakTestSuite) TestProcessing() {
	err := s.conn.Processing(s.jobUUID)
	assert.NoError(s.T(), err)

	i, err := s.conn.Status(s.jobUUID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, i.Processing)
}

func (s *RiakTestSuite) TestDone() {
	err := s.conn.Processing(s.jobUUID)
	assert.NoError(s.T(), err)

	err = s.conn.Processing(s.jobUUID)
	assert.NoError(s.T(), err)

	err = s.conn.Done(s.jobUUID)
	assert.NoError(s.T(), err)

	i, err := s.conn.Status(s.jobUUID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, i.Processing)
	assert.Equal(s.T(), 1, i.Done)
}

func (s *RiakTestSuite) TestSave() {
	err := s.conn.Save(s.jobUUID, "http://example.com/logo.png")
	assert.NoError(s.T(), err)

	r, err := s.conn.Results(s.jobUUID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(r))
}

func TestRiakSuite(t *testing.T) {
	if h, ok := context.ParseRiakHost(); ok {
		s := &RiakTestSuite{
			conn: context.ConnectRiakDb(h),
		}
		suite.Run(t, s)
	}
}
