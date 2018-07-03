all: clean install build

install:
	go install

build:
	$(GOPATH)/bin/go_homepage