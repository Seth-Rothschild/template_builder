install:
	./scripts/install.sh

format:
	go fmt ./...

lint:
	go vet ./...

test: lint format
	go test -v ./...

build:
	go build -o bin/template_builder .

e2e:
	./scripts/e2e-test.sh

run:
	go run .

ifeq (oneshot,$(firstword $(MAKECMDGOALS)))
  ONESHOT_FILE := $(word 2,$(MAKECMDGOALS))
  $(eval $(ONESHOT_FILE):;@:)
endif

oneshot: build
	./bin/template_builder $(ONESHOT_FILE)


