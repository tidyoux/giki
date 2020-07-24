GO_MODULE := github.com/tidyoux/giki
BIN_PATH := build/bin

LDFLAGS=-ldflags "-s -w"

all: server

server:
	cp -r static ${BIN_PATH}
	go build ${LDFLAGS} -o ${BIN_PATH}/giki ${GO_MODULE}

.PHONY: all server
