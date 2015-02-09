package db

import "github.com/tpjg/goriakpbc"

const (
	setsType     = "sets"
	mapsType     = "maps"
	countersType = "counters"

	jobsBucketKey        = "jobs"
	imagesSetKey         = "images"
	processingCounterKey = "processing"
	doneCounterKey       = "done"
	pageViewsKey         = "pagesView"

	objectNotFoundError = "Object not found"
)

// RiakConn implements the Connection interface using Riak as a backend.
// This is the prefered interface to use when running in a distributed environment.
type RiakConn struct {
	conn *riak.Client
	jobs *riak.Bucket
}

// NewRiakConn creates a new new instance of the database to talk with Riak.
// It assumes that the client has already been initialized by the application context.
func NewRiakConn(conn *riak.Client) (Connection, error) {
	j, err := conn.NewBucketType(mapsType, jobsBucketKey)
	if err != nil {
		return nil, err
	}

	return &RiakConn{
		conn: conn,
		jobs: j,
	}, nil
}

// Processing increments the counter of currently processing urls for a given job.
func (d RiakConn) Processing(jobUUID string) error {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return err
	}

	c := m.AddCounter(processingCounterKey)
	c.Increment(1)

	return m.Store()
}

// Done increments one element the counter of done urls
// and decrements the counter of processing urls for a given job.
func (d RiakConn) Done(jobUUID string) error {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return err
	}

	c := m.AddCounter(doneCounterKey)
	c.Increment(1)

	p := m.AddCounter(processingCounterKey)
	p.Increment(-1)

	return m.Store()
}

// Save adds an image source to the set of images for a given job.
func (d RiakConn) Save(jobUUID, src string) error {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return err
	}

	s := m.AddSet(imagesSetKey)
	s.Add([]byte(src))
	return m.Store()
}

// Status returns the processing and done counters of a given job.
func (d RiakConn) Status(jobUUID string) (*Info, error) {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return nil, err
	}

	info := &Info{}

	if c := m.FetchCounter(processingCounterKey); c != nil {
		info.Processing = c.GetValue()
	}

	if c := m.FetchCounter(doneCounterKey); c != nil {
		info.Done = c.GetValue()
	}

	var pages []Page
	if v := m.FetchMap(pageViewsKey); v != nil {
		for k := range v.Values {
			var pv int64
			if c := v.FetchCounter(k.Key); c != nil {
				pv = c.GetValue()
			}
			pages = append(pages, Page{k.Key, pv})
		}
	}
	info.pageViews = pages

	return info, nil
}

// Results returns the processed images for a given job.
func (d RiakConn) Results(jobUUID string) ([][]byte, error) {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return nil, err
	}

	var sv [][]byte
	if s := m.FetchSet(imagesSetKey); s != nil {
		sv = s.GetValue()
	}
	return sv, nil
}

// ViewPage decides whether a url needs to be visited or not.
// It assumed that you don't want to crawl the same url more than once
// in the current job.
func (d RiakConn) ViewPage(jobUUID string, url string) (bool, error) {
	m, err := d.jobs.FetchMap(jobUUID)
	if err != nil {
		return false, err
	}

	v := m.AddMap(pageViewsKey)
	c := v.AddCounter(url)

	view := c.GetValue() == 0
	c.Increment(1)
	err = m.Store()

	return view, err
}

// CreateJob initializes the job map in the Riak cluster.
// This operation must be performed before any crawling starts
// to guarantee that the process stores the data properly.
func (d RiakConn) CreateJob(jobUUID string) error {
	m := &riak.RDtMap{RDataTypeObject: riak.RDataTypeObject{Key: jobUUID, Bucket: d.jobs}}
	m.Init(nil)
	return m.Store()
}

func getCounter(bucket *riak.Bucket, jobUUID string) (int64, error) {
	c, err := bucket.FetchCounter(jobUUID)
	if err != nil {
		return 0, err
	}

	if c == nil {
		return 0, nil
	}

	return c.GetValue(), nil
}
