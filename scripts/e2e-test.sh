#!/bin/sh
set -e

port=8099
output=testdata/e2e-output.pdf

go build -o bin/template_builder .

PORT=$port ./bin/template_builder &
server_pid=$!
trap 'kill $server_pid 2>/dev/null' EXIT

sleep 1

if ! curl -fs "http://localhost:$port/" >/dev/null 2>&1; then
    echo "server did not become ready"
    exit 1
fi

curl -fs \
    -F "file=@testdata/e2e.md" \
    -F "images=@testdata/image.png" \
    "http://localhost:$port/build" \
    -o "$output"

if ! head -c 4 "$output" | grep -q "%PDF"; then
    echo "expected $output to be a pdf, got:"
    cat "$output"
    exit 1
fi

echo "e2e test passed: $output is a valid pdf"
