# Crawler

This is a distributed image crawler.

## Architecture

Crawler uses a distributed queue to publish messages with the URLs to crawl. Nodes subcribed to the queue pull the messages and crawl the URLs.
New URLs found crawling a page are sent to the queue for other nodes to process. Crawler has a hard limit of two layers depth crawling URLs, this means that it will crawl the URLs found in the first pages its receives but it will stop there (I don't really want to download the whole internet).

Crawler uses [Gnatsd](http://nats.io) as a clustered queue. New Nats server can be added to the cluster via configuration. New nodes can be subcribed to the queue by pointing the to cluster hosts.
Nats doesn't offer any durability guarantee, messages can be lost. If you need durability and deliveries guarantees you might want to take a look at [Vega](https://github.com/vektra/vega), a distributed
mailbox system that offers better delivery guarantees. See [Engines](#engines) to see how to swap queues implementations.

Crawler uses [Riak](https://github.com/basho/riak) as storage engine. Riak offers strong consistency guarantees using [CRDTs](http://pagesperso-systeme.lip6.fr/Marc.Shapiro/papers/RR-6956.pdf) and Crawler uses them to store counters and set of images processed.
The Riak engine can also be swapped if you want to use an engine with other kind of guarantees. See [Engines](#engines) to see how to swap storage implementations.

To bring new Crawler nodes up you just need to boot the application and point it to where Gnatsd and Riak are configured. See [Configuration](#configuration) next.

## Configuration

Crawler is configured via environment variables:

- CRAWLER_PORT: The port where the api is exposed, by default 3819. See [Api](#api) for more details about the api.
- CRAWLER_RIAK_URL: The host and port for one of the nodes to your Riak cluster, for instance `192.168.1.11:8087`. This is the port where the protocol buffers api is exposed in Riak.
- CRAWKER_GNATSD_NODES: The list of Gnatsd nodes configured separated by comma `,`, for instance `nats://192.168.1.12:4222,nats://192.168.1.13:4222`.

### Docker configuration

Crawler can be installed via [Docker](https://docker.com). You can pull the image `calavera/crawler` from the [Registry](https://registry.hub.docker.com/u/calavera/crawler).

Crawler can be linked with Riak and Gnatsd containers in your docker cluster. They need to be aliased as `riak` and `gnatsd` when you start the crawler container, for instance:

```
$ docker run --rm --link riak01:riak --link gnatsd:gnatsd -t calavera/crawler
```

## Api

Crawler exposes an http api at the port specified by `CRAWLER_PORT`. These are the enpoints that the api exposes:

- /: The root of the api can be reached via GET operations and displays a short howto about crawler.
- /crawl: This enpoint can be reached via POST to enqueue urls to crawl. The urls must be sent in the body of the request separated by white spaces, for instance:

```
$ curl -X POST -d@- http://localhost:3819/crawl << EOF
https://google.com https://cnn.com
EOF
```

When the messages are enqueued, this endpoint returns a job identifier in the body. It also sets the `Location` header to the status endpoint where you can check the status of the process.

- /status/job_uuid: This endpoint can be reached via GET. It displays the current urls processed, the ones that have been processed already and the number of times a urls is found in the crawling process.
- /results/job_uuid: This endpoint can be reached via GET. It displays the images that have been collected by the job.

## Engines

Crawler has been designed to be able to swap messaging and storage engines. In fact, you can see that it works if you start it without pointing it with the Gnatsd and Riak endpoints.
This is because, by default, Crawler starts in development mode with two **very unsafe** engines, a queue engine designed to use channels and a memory storage engine that is **not** thread safe.

There are two interfaces that you need to implement if you want to design new storage and messaging engines:

### Storage engines

The storage engines must implement this interface:

```go
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
```

### Messaging engines

The messaging engines must implement this interface:

```go
// Connection is an interface that defines how messages are published are received from a queue.
type Connection interface {
  // Publish pushes new messages to the queue.
  Publish(string, string, uint) error
  // Subscribe pulls messages from the queue and processes them using the processor function.
  Subscribe(Processor)
}
```

## Building

The build system assumes you're in a Linux host and you have Go and Docker installed. Run `make build` to generate a docker container.

## TL;DR

Run `make tldr` if you want to start right away to test it. This task assumes that you have curl and a Docker client installed and connected with a Docker host.

- It pulls `apcera/gnatsd` from the Docker registry and starts a node.
- It pulls `hectcastro/riak` from the Docker registry and starts a cluster of 5 nodes.
- It pulls `calavera/crawler` from the Docker registry and starts two nodes.
- It sends an initial request to crawl `https://google.com` and `https://docker.com`.

## License

MIT
