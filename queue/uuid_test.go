package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	uuid := UUID()

	assert.NotNil(t, uuid)
	assert.NotEqual(t, "", uuid)
}
