package crawler

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/PuerkitoBio/fetchbot"
	"github.com/PuerkitoBio/goquery"
	"github.com/calavera/crawler/db"
	"github.com/calavera/crawler/queue"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func TestCrawlDocuments(t *testing.T) {
	testCases := []struct {
		id     string
		page   string
		images int
	}{
		{
			id:     "test1",
			page:   "simple_page.html",
			images: 1,
		},
		{
			id:     "test2",
			page:   "multi_images.html",
			images: 2,
		},
		{
			id:     "test3",
			page:   "empty_page.html",
			images: 0,
		},
	}

	for _, e := range testCases {
		d, _ := db.NewMapConn()
		c := newCrawler(d, queue.NewPoolConn(d), queue.NewMessage(e.id, "http://example.com", 1))

		doc := loadPage(t, e.page)
		x := loadContext(t, "http://example.com")
		c.crawlDocument(x, doc)

		r, _ := d.Results(e.id)
		assert.Equal(t, e.images, len(r))
	}
}

func TestContinueCrawling(t *testing.T) {
	d, _ := db.NewMapConn()
	c := newCrawler(d, queue.NewPoolConn(d), queue.NewMessage("test", "http://example.com", 1))
	assert.False(t, c.continueCrawling())

	c = newCrawler(d, queue.NewPoolConn(d), queue.NewMessage("test", "http://example.com", 0))
	assert.True(t, c.continueCrawling())
}

func TestCrawlHref(t *testing.T) {
	d, _ := db.NewMapConn()
	p := queue.NewPoolConn(d)
	c := newCrawler(d, p, queue.NewMessage("test", "http://example.com", 0))
	x := loadContext(t, "http://example.com")

	done := make(chan bool)
	processor := func(q queue.Connection, d db.Connection, msg *queue.Message) {
		doc := loadPage(t, "simple_page.html")
		c.crawlDocument(x, doc)
		done <- true
	}

	p.Subscribe(processor)

	doc := loadPage(t, "follow_index.html")
	c.crawlDocument(x, doc)

	<-done
	r, _ := d.Results("test")
	assert.Equal(t, 1, len(r))

	assert.Equal(t, "http://example.com/images/logo.jpg", string(r[0]))
}

func loadContext(t *testing.T, s string) *fetchbot.Context {
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	cmd := &fetchbot.Cmd{
		U: u,
		M: "GET",
	}

	return &fetchbot.Context{
		Cmd: cmd,
	}
}

func loadPage(t *testing.T, page string) *goquery.Document {
	var f *os.File
	var e error

	if f, e = os.Open(fmt.Sprintf("./test_data/%s", page)); e != nil {
		t.Fatal(e)
	}
	defer f.Close()

	var node *html.Node
	if node, e = html.Parse(f); e != nil {
		t.Fatal(e)
	}
	return goquery.NewDocumentFromNode(node)
}
