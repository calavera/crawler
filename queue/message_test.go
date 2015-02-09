package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	m := NewMessage("test", "http://example.com", 0)

	assert.Equal(t, "test", m.JobUUID)
	assert.Equal(t, "http://example.com", m.URL)
	assert.Equal(t, 0, m.Depth)
}
