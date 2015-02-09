#! /bin/sh

set -e

if [ -z "${DOCKER_HOST}" ]; then
  echo ""
  echo "It looks like the environment variable DOCKER_HOST has not"
  echo "been set.  The Riak cluster cannot be started unless this has"
  echo "been set appropriately.  For example:"
  echo ""
  echo "  export DOCKER_HOST=\"tcp://127.0.0.1:2375\""
  echo ""

  exit 1
fi

if [[ "${DOCKER_HOST}" == unix://* ]]; then
  CLEAN_DOCKER_HOST="localhost"
else
  CLEAN_DOCKER_HOST=$(echo "${DOCKER_HOST}" | cut -d'/' -f3 | cut -d':' -f1)
fi

docker pull hectcastro/riak

scripts/start-riak.sh

echo
echo "Bringing up Gnatsd nodes:"
echo
docker run -d --name gnatsd -p 4222:4222 apcera/gnatsd


echo
echo "Bringing up Crawler nodes:"
echo
docker run -d --name crawler01 --link gnatsd:gnatsd --link riak01:riak -p 3819:3819 -t calavera/crawler
docker run -d --name crawler02 --link gnatsd:gnatsd --link riak01:riak -p 3820:3820 -e "CRAWLER_PORT=3820" -t calavera/crawler

echo
echo "Waiting a second until the Crawler cluster starts"
echo
sleep 1

JOB_UUID=`curl -s -X POST -d@- http://${CLEAN_DOCKER_HOST}:3819/crawl << EOF
https://google.com https://www.docker.com
EOF`

echo
echo "URLs sent for crawling you can see the process and results in these urls:"
echo
echo "\t- Status:  http://${CLEAN_DOCKER_HOST}:3819/status/${JOB_UUID}"
echo "\t- Results: http://${CLEAN_DOCKER_HOST}:3819/results/${JOB_UUID}"
echo
