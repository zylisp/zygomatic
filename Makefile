VERSION_SRC = src/github.com/zylisp/zylisp/gitcommit.go
LAST_TAG = $(shell git describe --abbrev=0 --tags)
LAST_COMMIT = $(shell git rev-parse --short HEAD)

.PHONY: build test all

all: build test

deps:
	go get github.com/glycerine/blake2b
	go get github.com/glycerine/greenpack/msgp
	go get github.com/glycerine/liner
	go get github.com/shurcooL/go-goon
	go get github.com/tinylib/msgp/msgp
	go get github.com/ugorji/go/codec

lint-deps:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | \
	sh -s -- -b ~/go/bin v1.15.0

test-deps:
	go get github.com/glycerine/goconvey/convey

build: deps
	/bin/echo "package zylisp" > $(VERSION_SRC)
	/bin/echo "" >> $(VERSION_SRC)
	/bin/echo "func init() { GITLASTTAG = \"$(LAST_TAG)\"; \
	GITLASTCOMMIT = \"$(LAST_COMMIT)\" }" >> $(VERSION_SRC)
	go install github.com/zylisp/zylisp/cmd/zylisp

lint: lint-deps
	golangci-lint run

vet:
	go vet github.com/zylisp/zcore
	go vet github.com/zylisp/zylisp

test: test-deps
	tests/testall.sh && \
	echo "running 'go test'" && \
	cd src/github.com/zylisp/zcore && \
	go test -v
	# cd - && \
	# cd src/github.com/zylisp/zylisp && \
	# go test -v

