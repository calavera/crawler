.PHONY: deps test build tldr

test:
	go test ./...

build:
	mkdir -p bin
	GOPATH=$(GOPATH):$(GOPATH)/src/github.com/calavera/crawler/Godeps/_workspace CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o bin/crawler ./cmd
	docker build -t calavera/crawler .

tldr:
	scripts/tldr.sh
