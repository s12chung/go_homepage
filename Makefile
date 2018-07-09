.PHONY: server

all: install build

build: build-assets build-go

dev: clean build
	make server & make watch-go & make watch-assets

# See https://github.com/webpack/webpack/issues/2537#issuecomment-280447557
prod: clean
	NODE_ENV=production webpack -p
	$(GOPATH)/bin/go_homepage
	make server

install:
	dep ensure
	go install
	yarn install --pure-lockfile

docker:
	docker-compose up

docker-rm:
	docker-compose rm -v -s

clean:
	rm -rf generated node_modules/.cache/

server:
	$(GOPATH)/bin/go_homepage -server

build-go:
	$(GOPATH)/bin/go_homepage

build-assets:
	webpack

# https://github.com/brandur/sorg/blob/28ac85ff5fd6caf57da974aff2a1af97f8943ef3/Makefile#L132
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./vendor/*")
TMPL_FILES := $(shell find ./templates -type f -name "*.tmpl")
MARKDOWN_FILES = $(shell find ./assets/markdowns -type f -name "*.md")
watch-go:
	fswatch -o $(GO_FILES) $(TMPL_FILES) $(MARKDOWN_FILES) vendor/ generated/assets/manifest.json | xargs -n1 -I{} make install build-go

SCSS_FILES := $(shell find . -type f -name "*.scss" ! -path "./node_modules/*")
JS_FILES := $(shell find . -type f -name "*.js" ! -path "./node_modules/*" ! -path "./generated/*")
watch-assets:
	fswatch -v -o $(SCSS_FILES) $(JS_FILES) | xargs -n1 -I{} make build-assets
