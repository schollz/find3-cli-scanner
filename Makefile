.PHONY: scratch, install, basicbuild, server, server1, server2, server3, dev1, dev2, dev3


HASH=$(shell git describe --tags)
LDFLAGS=-ldflags "-s -w -X main.version=${HASH}"

basicbuild: getmage
	mage

release: basicbuild
	mage release

install: 
	mv find3-cli-scanner /usr/local/bin/

getmage:
	go get github.com/magefile/mage 
