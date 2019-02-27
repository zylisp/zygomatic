VERSION_SRC=src/github.com/zylisp/zygo/gitcommit.go

.PHONY: build test all

all: build test

deps:
	go get github.com/glycerine/blake2b
	go get github.com/glycerine/greenpack/msgp
	go get github.com/glycerine/liner
	go get github.com/shurcooL/go-goon
	go get github.com/tinylib/msgp/msgp
	go get github.com/ugorji/go/codec

test-deps:
	go get github.com/glycerine/goconvey/convey

build: deps
	/bin/echo "package zygo" > $(VERSION_SRC)
	/bin/echo "func init() { GITLASTTAG = \"$(shell git describe --abbrev=0 --tags)\"; \
	GITLASTCOMMIT = \"$(shell git rev-parse HEAD)\" }" >> $(VERSION_SRC)
	go install github.com/zylisp/zygo/cmd/zygo

test: test-deps
	tests/testall.sh && \
	echo "running 'go test'" && \
	cd src/github.com/zylisp/zygo && \
	go test -v
