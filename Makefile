SHORT_TTL := 3600
# for assets, which don't change or have hashed filenames
LONG_TTL := 86400

.PHONY: server

all: install build

install:
	dep ensure
	yarn install --pure-lockfile

build: build-assets build-go

dev: clean build watch watch-install
	make -j file-server watch-logs

# See https://github.com/webpack/webpack/issues/2537#issuecomment-280447557
prod: clean
	NODE_ENV=production webpack -p
	go install
	$(GOPATH)/bin/go_homepage

test:
	go test ./go/...

push-docker-deploy:
	git push origin master
	make docker-deploy

deploy:
	aws s3 sync $(GENERATED_PATH) s3://$(S3_BUCKET)/ --cache-control max-age=$(SHORT_TTL) --delete --content-type text/html --exclude '$(ASSETS_PATH)/*' --exclude '*.*' --include '*.html'
	aws s3 sync $(GENERATED_PATH)/$(ASSETS_PATH) s3://$(S3_BUCKET)/$(ASSETS_PATH)/ --cache-control max-age=$(LONG_TTL) --delete
	aws s3 cp $(GENERATED_PATH)/favicon.ico s3://$(S3_BUCKET)/ --cache-control max-age=$(LONG_TTL) --content-type image/x-icon
	aws s3 cp $(GENERATED_PATH)/browserconfig.xml s3://$(S3_BUCKET)/ --cache-control max-age=$(LONG_TTL) --content-type application/xml
	aws s3 cp $(GENERATED_PATH)/posts.atom s3://$(S3_BUCKET)/ --cache-control max-age=$(SHORT_TTL) --content-type application/xml
	aws s3 cp $(GENERATED_PATH)/robots.txt s3://$(S3_BUCKET)/ --cache-control max-age=$(SHORT_TTL) --content-type text/plain

docker-install: docker-build-install docker-copy

docker-build-install:
	docker-compose up --no-start

# $(shell docker-compose ps -q web) breaks if this target is combined with docker-build
docker-copy:
	docker cp $(shell docker-compose ps -q web):$(DOCKER_WORKDIR)/node_modules ./node_modules
	docker cp $(shell docker-compose ps -q web):$(DOCKER_WORKDIR)/vendor ./vendor

docker-build:
	docker-compose up --build --no-start

docker:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

docker-prod:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up

docker-deploy:
	docker-compose -f docker-compose.yml -f docker-compose.deploy.yml up

docker-sh:
	docker-compose exec web ash

docker-rm:
	docker-compose rm -v -s

clean:
	rm -rf $(GENERATED_PATH) ./node_modules/.cache/
	watchman watch-del-all
	watchman shutdown-server

clean-all: clean
	rm -rf cache

server: clean build-assets
	go install
	$(GOPATH)/bin/go_homepage -server

file-server:
	$(GOPATH)/bin/go_homepage -file-server

build-go:
	go install
	$(GOPATH)/bin/go_homepage

build-assets:
	webpack --color

watch:
	watchman watch-project .

watch-logs:
	mkdir -p logs
	touch logs/watchman-build.log
	tail -f logs/watchman-build.log

watch-install:
	watchman -j < watchman/build-go.json
	watchman -j < watchman/build-assets.json