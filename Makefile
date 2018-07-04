all: clean install build

install:
	go install

build:
	$(GOPATH)/bin/go_homepage

clean:
	rm -rf generated

# https://github.com/brandur/sorg/blob/28ac85ff5fd6caf57da974aff2a1af97f8943ef3/Makefile#L132
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./vendor/*")
watch:
	make install build
	fswatch -o $(GO_FILES) vendor/ | xargs -n1 -I{} make install build