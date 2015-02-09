package db

import (
	"fmt"
	"sync"
)

// set is a very inneficient memory set designed for testing.
type set struct {
	*sync.Mutex
	keys   map[string]string
	values [][]byte
}

// Values returns the list of values in the set.
func (s *set) Values() [][]byte {
	s.Lock()
	defer s.Unlock()

	return s.values
}

// Add appends new values to the set.
func (s *set) Add(el string) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.keys[el]; !ok {
		s.keys[el] = el
		s.values = append(s.values, []byte(el))
	}
}

func newSet() *set {
	return &set{
		Mutex:  new(sync.Mutex),
		keys:   map[string]string{},
		values: make([][]byte, 0),
	}
}

// MapConn implements the Connection interface using memory maps as backends.
// This interface is only suitable for testing.
// It offers no guarantees about the elements saved in it and it is not thread safe.
type MapConn struct {
	images     map[string]*set
	processing map[string]int64
	done       map[string]int64
	pageViews  map[string]map[string]int64
}

// NewMapConn creates a new map connection.
func NewMapConn() (Connection, error) {
	return &MapConn{
		images:     map[string]*set{},
		processing: map[string]int64{},
		done:       map[string]int64{},
		pageViews:  map[string]map[string]int64{},
	}, nil
}

// Processing increments the counter of current urls processing.
func (c *MapConn) Processing(jobUUID string) error {
	var p int64
	if v, ok := c.processing[jobUUID]; ok {
		p = v
	}
	c.processing[jobUUID] = p + 1

	return nil
}

// Done increments the counter of urls processed
// and decrements the counter of urls currently processing.
func (c *MapConn) Done(jobUUID string) error {
	var p int64
	if v, ok := c.done[jobUUID]; ok {
		p = v
	}
	c.done[jobUUID] = p + 1

	if v, ok := c.processing[jobUUID]; ok {
		c.processing[jobUUID] = v - 1
	}

	return nil
}

// Save stores new images found by a job in the database.
func (c *MapConn) Save(jobUUID string, src string) error {
	set := newSet()
	if s, ok := c.images[jobUUID]; ok {
		set = s
	}
	set.Add(src)
	c.images[jobUUID] = set
	return nil
}

// Status gives you information about the current job.
// It returns the currently processing urls and the urls already processed.
// It also returns the urls detected by the job.
func (c *MapConn) Status(jobUUID string) (*Info, error) {
	var c1 int64
	var ok bool

	if c1, ok = c.processing[jobUUID]; !ok {
		return nil, fmt.Errorf("job not found") // nothing processed yet
	}

	c2 := c.done[jobUUID]

	var pages []Page
	if c.pageViews[jobUUID] != nil {
		for k, v := range c.pageViews[jobUUID] {
			pages = append(pages, Page{k, v})
		}
	}

	return &Info{
		Processing: c1,
		Done:       c2,
		pageViews:  pages,
	}, nil
}

// Results returns the list of images crawled by a specific job.
func (c *MapConn) Results(jobUUID string) ([][]byte, error) {
	if s, ok := c.images[jobUUID]; ok {
		return s.Values(), nil
	}
	return nil, fmt.Errorf("job not found")
}

// ViewPage decides whether a url needs to be visited or not.
// It assumed that you don't want to crawl the same url more than once
// in the current job.
func (c *MapConn) ViewPage(jobUUID string, url string) (bool, error) {
	if _, ok := c.pageViews[jobUUID]; !ok {
		c.pageViews[jobUUID] = map[string]int64{}
	}

	if _, ok := c.pageViews[jobUUID][url]; ok {
		c.pageViews[jobUUID][url]++
		return false, nil
	}

	c.pageViews[jobUUID][url] = 1
	return true, nil
}

// CreateJob is NOOP in the map storage.
func (c *MapConn) CreateJob(jobUUID string) error {
	return nil
}
