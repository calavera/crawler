# Functional tests

This directory include test cases that require of an external service to work.

## Riak tests

1. Set the Riak host via the `CRAWLER_RIAK_URL` env variable before running the funtional tests.
2. Run `make test`

## Gnatsd tests

1. Set the Gnatsd nodes via the `CRAWLER_GNATSD_NODES` env variable before running the funtional tests.
2. Run `make test`
