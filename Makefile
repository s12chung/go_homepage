all: install build

install:
	go install

build:
	webpack
	$(GOPATH)/bin/go_homepage

production: clean
	# so NODE_ENV=production sets webpack env, -p sets the compiled JS env, https://github.com/webpack/webpack/issues/2537#issuecomment-280447557
	NODE_ENV=production webpack -p
	$(GOPATH)/bin/go_homepage

clean:
	rm -rf generated node_modules/.cache/

# https://github.com/brandur/sorg/blob/28ac85ff5fd6caf57da974aff2a1af97f8943ef3/Makefile#L132
GO_FILES := $(shell find . -type f -name "*.go" ! -path "./vendor/*")
watch:
	make install build
	fswatch -o $(GO_FILES) vendor/ | xargs -n1 -I{} make install build