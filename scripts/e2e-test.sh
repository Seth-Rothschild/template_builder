#!/bin/sh
set -e

port=8099
output=testdata/e2e-output.pdf

go build -o bin/template_builder .

PORT=$port ./bin/template_builder &
server_pid=$!
trap 'kill $server_pid 2>/dev/null' EXIT

sleep 1

status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port/health")
if [ "$status" != "200" ]; then
    echo "expected /health to return 200, got $status"
    exit 1
fi

echo "e2e test passed: /health returned 200"

oneshot_dir=$(mktemp -d)
cp testdata/minimal.md "$oneshot_dir/minimal.md"
make oneshot "$oneshot_dir/minimal.md"

if ! head -c 4 "$oneshot_dir/minimal.pdf" | grep -q "%PDF"; then
    echo "expected $oneshot_dir/minimal.pdf to be a pdf"
    exit 1
fi

echo "e2e test passed: make oneshot produced a valid pdf"

curl -fs \
    --data-binary @testdata/minimal.md \
    "http://localhost:$port/build" \
    -o "$output"

if ! head -c 4 "$output" | grep -q "%PDF"; then
    echo "expected $output to be a pdf, got:"
    cat "$output"
    exit 1
fi

echo "e2e test passed: $output is a valid pdf"
