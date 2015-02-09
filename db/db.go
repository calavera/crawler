package db

import "sort"

// Connection is an interface that defines how data is saved and retrieved from a storage.
type Connection interface {
	// Create the job in the database.
	CreateJob(string) error
	// Processing increments the counter of currently processing urls for a given job.
	Processing(string) error
	// Done increments the counter of done urls
	// and decrements the counter of processing urls for a given job.
	Done(string) error
	// Save adds an image source to the set of images for a given job.
	Save(string, string) error
	// Status returns the processing and done counters of a given job.
	Status(string) (*Info, error)
	// Results returns the processed images for a given job.
	Results(string) ([][]byte, error)
	// ViewPage decides whether a page needs to be crawled or not.
	// One url should only be crawled once by a given job,
	// but it depends on the guarantees that the storage provides.
	ViewPage(string, string) (bool, error)
}

// Page represents a visited url.
// It stores how many times a job has seen the page.
type Page struct {
	URL  string
	Hits int64
}

// Pages is a sortable collection of pages.
type Pages []Page

func (p Pages) Len() int {
	return len(p)
}

func (p Pages) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Pages) Less(i, j int) bool {
	return p[i].Hits >= p[j].Hits
}

// Info stores information about a specific job.
type Info struct {
	Processing int64
	Done       int64
	pageViews  []Page
}

// PageViews returns the urls found by a specific job
// in descending order by the number or occurrences.
func (i *Info) PageViews() Pages {
	sort.Sort(Pages(i.pageViews))
	return i.pageViews
}
