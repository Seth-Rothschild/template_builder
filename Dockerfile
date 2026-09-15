FROM golang:1.25 AS build

WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY builder ./builder
RUN CGO_ENABLED=0 go build -o /template_builder .

FROM debian:bookworm-slim

ARG TARGETARCH
ARG PANDOC_VERSION=3.11

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL -o /tmp/pandoc.deb "https://github.com/jgm/pandoc/releases/download/${PANDOC_VERSION}/pandoc-${PANDOC_VERSION}-1-${TARGETARCH}.deb" \
    && dpkg -i /tmp/pandoc.deb \
    && rm /tmp/pandoc.deb

RUN curl --proto '=https' --tlsv1.2 -fsSL https://drop-sh.fullyjustified.net | sh \
    && mv ./tectonic /usr/local/bin/tectonic

ENV TECTONIC_CACHE_DIR=/opt/tectonic-cache

WORKDIR /app
COPY --from=build /template_builder /app/template_builder
COPY templates ./templates
COPY filters ./filters

COPY testdata/minimal.md testdata/e2e.md testdata/image.png /tmp/warm/
RUN ./template_builder /tmp/warm/minimal.md \
    && ./template_builder /tmp/warm/e2e.md \
    && rm -rf /tmp/warm

EXPOSE 8080
ENTRYPOINT ["/app/template_builder"]
