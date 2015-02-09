package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortedPageViews(t *testing.T) {
	pages := []Page{
		Page{"http://example.com", 1},
		Page{"http://example.org", 5},
		Page{"http://example.net", 2},
	}

	info := &Info{pageViews: pages}

	s := info.PageViews()
	assert.Equal(t, "http://example.org", s[0].URL)
	assert.Equal(t, "http://example.net", s[1].URL)
	assert.Equal(t, "http://example.com", s[2].URL)
}
