.PHONY: build build-image build-cli test clean

HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

build: build-cli build-image

build-cli:
	go build -o bin/agx ./cmd/agx

build-image:
	docker build \
	  --build-arg HOST_UID=$(HOST_UID) \
	  --build-arg HOST_GID=$(HOST_GID) \
	  -t agx:latest image/

test:
	go test ./...

clean:
	docker image rm agx:latest 2>/dev/null || true
	rm -rf bin/
