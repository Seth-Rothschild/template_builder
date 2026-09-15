install:
	./scripts/install.sh

format:
	go fmt ./...

lint:
	go vet ./...

test: lint format
	for f in filters/*_test.lua; do pandoc lua "$$f"; done
	go test -v ./...
	./scripts/e2e-test.sh

benchmark:
	go test -run=^$$ -bench=BenchmarkBuildHandler$$ -benchtime=10x .
	go test -run=^$$ -bench=BenchmarkBuildHandlerParallel -benchtime=16x -cpu=1,2,4,8 .

build:
	go build -o bin/template_builder .

run:
	go run .

ifeq (oneshot,$(firstword $(MAKECMDGOALS)))
  ONESHOT_FILE := $(word 2,$(MAKECMDGOALS))
  $(eval $(ONESHOT_FILE):;@:)
endif

oneshot: build
	./bin/template_builder $(ONESHOT_FILE)


