FROM scratch
MAINTAINER David Calavera <david.calavera@gmail.com>

COPY bin/crawler /
ENTRYPOINT ["/crawler"]
