package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/calavera/crawler/context"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	d, _ := db.NewMapConn()
	x := context.Context{d, nil}

	s := newServer(x)

	r, _ := http.NewRequest("GET", "http://example.com", nil)

	p := make(httprouter.Params, 1)
	p[0].Key = "jobUUID"
	p[0].Value = "test"

	w := httptest.NewRecorder()
	s.status(w, r, p)
	assert.Equal(t, 404, w.Code)

	w = httptest.NewRecorder()
	s.results(w, r, p)
	assert.Equal(t, 404, w.Code)
}

func TestFound(t *testing.T) {
	d, _ := db.NewMapConn()
	d.ViewPage("test", "http://example.com")
	d.Processing("test")
	d.Save("test", "http://example.com/image.jpg")

	x := context.Context{d, nil}

	s := newServer(x)

	r, _ := http.NewRequest("GET", "http://example.com", nil)

	p := make(httprouter.Params, 1)
	p[0].Key = "jobUUID"
	p[0].Value = "test"

	w := httptest.NewRecorder()
	s.status(w, r, p)
	assert.Equal(t, 200, w.Code)

	b, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, "- Processing: 1 URLs\n- Done: 0 URLs\n- Page views:\n\t- http://example.com -> 1 hit", string(b))

	w = httptest.NewRecorder()
	s.results(w, r, p)
	assert.Equal(t, 200, w.Code)

	b, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/image.jpg\n", string(b))
}

func TestParseURLs(t *testing.T) {
	testCases := []struct {
		input string
		urls  int
	}{
		{
			input: "http://example.com",
			urls:  1,
		},
		{
			input: "http://example.com\nhttp://example2.com",
			urls:  2,
		},
		{
			input: "http://example.com http://example2.com",
			urls:  2,
		},
		{
			input: "",
			urls:  0,
		},
	}

	for _, tc := range testCases {
		r, _ := http.NewRequest("GET", "http://example.com", strings.NewReader(tc.input))

		urls, err := parseURLs(r)
		assert.NoError(t, err)
		assert.Equal(t, tc.urls, len(urls))
	}
}

func TestCrawl(t *testing.T) {
	d, _ := db.NewMapConn()
	q := queue.NewPoolConn(d)

	counter := 0
	processor := func(q queue.Connection, d db.Connection, msg *queue.Message) {
		counter++
	}
	q.Subscribe(processor)

	x := context.Context{d, q}

	s := newServer(x)
	r, _ := http.NewRequest("GET", "http://example.com", strings.NewReader("http://example.com"))

	p := make(httprouter.Params, 0)
	w := httptest.NewRecorder()
	s.crawl(w, r, p)

	b, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	j := string(b)
	assert.Equal(t, fmt.Sprintf("/status/%s", j), w.Header().Get("Location"))

	assert.Equal(t, 201, w.Code)
}

func TestIndex(t *testing.T) {
	x := context.Context{}
	s := newServer(x)

	r, _ := http.NewRequest("GET", "http://example.com", nil)
	p := make(httprouter.Params, 0)

	w := httptest.NewRecorder()
	s.index(w, r, p)
	assert.Equal(t, 200, w.Code)

	b, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.Equal(t, usage, string(b))
}

func TestServerPort(t *testing.T) {
	p := serverPort()
	assert.Equal(t, ":3819", p)

	os.Setenv("CRAWLER_PORT", "3820")
	defer os.Setenv("CRAWLER_PORT", "")

	p = serverPort()
	assert.Equal(t, ":3820", p)
}

func TestCreateJob(t *testing.T) {
	d, _ := db.NewMapConn()
	x := context.Context{d, nil}

	s := newServer(x)
	j, err := s.createNewJob()
	assert.NoError(t, err)
	assert.NotNil(t, j)
}
