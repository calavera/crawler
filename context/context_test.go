package context

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfiguration(t *testing.T) {
	prevRiak := os.Getenv(riakAddressKey)
	prevNats := os.Getenv(natsNodesKey)
	os.Setenv(riakAddressKey, "")
	os.Setenv(natsNodesKey, "")

	defer os.Setenv(riakAddressKey, prevRiak)
	defer os.Setenv(natsNodesKey, prevNats)

	_, ok := ParseRiakHost()
	assert.False(t, ok)

	_, ok = ParseNatsNodes()
	assert.False(t, ok)
}

func TestRiakHostConfiguration(t *testing.T) {
	prevRiak := os.Getenv(riakAddressKey)
	os.Setenv(riakAddressKey, "192.168.59.103:8087")
	defer os.Setenv(riakAddressKey, prevRiak)

	h, ok := ParseRiakHost()
	assert.True(t, ok)
	assert.Equal(t, "192.168.59.103:8087", h)
}

func TestGnatsClusterConfiguration(t *testing.T) {
	prevNats := os.Getenv(natsNodesKey)
	os.Setenv(natsNodesKey, "nats://localhost:1222, nats://localhost:1223")
	defer os.Setenv(natsNodesKey, prevNats)

	s, ok := ParseNatsNodes()
	assert.True(t, ok)
	assert.Equal(t, []string{"nats://localhost:1222", "nats://localhost:1223"}, s)
}

func TestRiakDockerLink(t *testing.T) {
	os.Setenv(riakDockerLinkKey, "tcp://192.168.59.103:8087")
	defer os.Setenv(riakDockerLinkKey, "")

	h, ok := ParseRiakHost()
	assert.True(t, ok)
	assert.Equal(t, "192.168.59.103:8087", h)
}

func TestGnatsdDockerLink(t *testing.T) {
	os.Setenv(natsDockerLinkKey, "tcp://192.168.59.103:1223")
	defer os.Setenv(natsDockerLinkKey, "")

	h, ok := ParseNatsNodes()
	assert.True(t, ok)
	assert.Equal(t, []string{"nats://192.168.59.103:1223"}, h)
}
