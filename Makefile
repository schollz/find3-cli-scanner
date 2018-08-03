.PHONY: scratch, install, basicbuild, server, server1, server2, server3, dev1, dev2, dev3


HASH=$(shell git describe --tags)
LDFLAGS=-ldflags "-s -w -X main.version=${HASH}"

basicbuild:
	# make sure you have libpcap
	# Linux: apt-get install libpcap-dev
	# OS X: brew install libpcap
	go build -v ${LDFLAGS}

