#!/bin/sh
set -e

port=8099
output=testdata/e2e.pdf

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
cp testdata/e2e.md "$oneshot_dir/e2e.md"
make oneshot "$oneshot_dir/e2e.md"

if ! head -c 4 "$oneshot_dir/e2e.pdf" | grep -q "%PDF"; then
    echo "expected $oneshot_dir/e2e.pdf to be a pdf"
    exit 1
fi

echo "e2e test passed: make oneshot produced a valid pdf"

curl -fs \
    --data-binary @testdata/e2e.md \
    "http://localhost:$port/build" \
    -o "$output"

if ! head -c 4 "$output" | grep -q "%PDF"; then
    echo "expected $output to be a pdf, got:"
    cat "$output"
    exit 1
fi

echo "e2e test passed: $output is a valid pdf"

multipart_output=testdata/e2e-multipart.pdf

curl -fs \
    -F markdown="<testdata/e2e.md" \
    -F images=@testdata/image.png \
    "http://localhost:$port/build" \
    -o "$multipart_output"

if ! head -c 4 "$multipart_output" | grep -q "%PDF"; then
    echo "expected $multipart_output to be a pdf, got:"
    cat "$multipart_output"
    exit 1
fi

if ! grep -aq "/Image" "$multipart_output"; then
    echo "expected $multipart_output to embed the uploaded image"
    exit 1
fi

echo "e2e test passed: $multipart_output embeds the uploaded image"
