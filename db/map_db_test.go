package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	s := newSet()

	s.Add("test")
	assert.Equal(t, 1, len(s.Values()))

	s.Add("done")
	assert.Equal(t, 2, len(s.Values()))

	s.Add("test")
	assert.Equal(t, 2, len(s.Values()))
}

func TestMapDbProcessing(t *testing.T) {
	m, _ := NewMapConn()
	m.Processing("test")

	s, _ := m.Status("test")
	assert.Equal(t, 1, s.Processing)
}

func TestMapDbDone(t *testing.T) {
	m, _ := NewMapConn()
	m.Processing("test")
	m.Done("test")

	s, _ := m.Status("test")
	assert.Equal(t, 0, s.Processing)
	assert.Equal(t, 1, s.Done)

	m.Processing("test")
	m.Processing("test")

	m.Done("test")
	s, _ = m.Status("test")
	assert.Equal(t, 1, s.Processing)
	assert.Equal(t, 2, s.Done)
}

func TestMapDbSave(t *testing.T) {
	m, _ := NewMapConn()

	_, err := m.Results("test")
	assert.Error(t, err)

	m.Save("test", "src1")
	m.Save("test", "src2")

	i, err := m.Results("test")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(i))

	m.Save("test2", "src")

	i, err = m.Results("test2")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(i))
}

func TestMapDbViewPage(t *testing.T) {
	m, _ := NewMapConn()
	m.Processing("test")

	v, _ := m.ViewPage("test", "test")
	assert.True(t, v)

	v, _ = m.ViewPage("test", "test")
	assert.False(t, v)

	v, _ = m.ViewPage("test", "test")
	assert.False(t, v)

	s, _ := m.Status("test")
	assert.Equal(t, 3, s.PageViews()[0].Hits)
}
