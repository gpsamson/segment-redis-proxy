IMAGE ?= gpsamson/segment-redis-proxy
TAG ?= $(shell git rev-parse --short HEAD)

container-image:
	docker build -t $(IMAGE):$(TAG) .

check: format lint test

format:
	go fmt ./...

lint:
	@if ! which gometalinter &>/dev/null; then \
	    echo Installing gometalinter...; \
	    go get -u github.com/alecthomas/gometalinter; \
	    gometalinter --install; \
	fi
	gometalinter --tests --vendor --aggregate --enable-gc --disable=gocyclo --deadline=5m ./...

test:
	go test ./... -v

deps:
	@which dep &>/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure

.PHONY: docker-image format lint test deps
