.PHONY: server

all: install build

install:
	dep ensure
	yarn install --pure-lockfile

build: build-assets build-go

dev: clean build watch watch-install
	make -j server watch-logs

# See https://github.com/webpack/webpack/issues/2537#issuecomment-280447557
prod: clean
	NODE_ENV=production webpack -p
	$(GOPATH)/bin/go_homepage

docker:
	docker-compose up

docker-rm:
	docker-compose rm -v -s

clean:
	rm -rf generated node_modules/.cache/
	watchman watch-del-all
	watchman shutdown-server

server:
	$(GOPATH)/bin/go_homepage -server

build-go:
	go install
	$(GOPATH)/bin/go_homepage

build-assets:
	webpack

watch:
	watchman watch-project .

watch-logs:
	touch logs/watchman-build.log
	tail -f logs/watchman-build.log

watch-install:
	watchman -j < watchman/build-go.json
	watchman -j < watchman/build-assets.json