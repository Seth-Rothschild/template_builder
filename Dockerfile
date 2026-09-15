FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
COPY builder ./builder
RUN CGO_ENABLED=0 go build -o /template_builder .

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl pandoc \
    && rm -rf /var/lib/apt/lists/*
RUN curl --proto '=https' --tlsv1.2 -fsSL https://drop-sh.fullyjustified.net | sh \
    && mv ./tectonic /usr/local/bin/tectonic

ENV TECTONIC_CACHE_DIR=/opt/tectonic-cache
COPY testdata/sample.tex /tmp/warm/sample.tex
RUN tectonic /tmp/warm/sample.tex --outdir /tmp/warm \
    && rm -rf /tmp/warm

COPY --from=build /template_builder /template_builder
EXPOSE 8080
ENTRYPOINT ["/template_builder"]
