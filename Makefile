install:
	./scripts/install.sh

format:
	go fmt ./...

lint:
	go vet ./...

test: lint format
	go clean -testcache
	go test --race -v ./...
	pandoc lua filters/tables_test.lua

build:
	go build -o bin/template_builder .

e2e:
	./scripts/e2e-test.sh

run:
	go run .


