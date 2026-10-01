FROM golang:1.25 AS builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED=1
ARG VERSION=dev
RUN go build -o bin/qq -ldflags="-linkmode external -extldflags -static -X github.com/JFryy/qq/cli.Version=${VERSION}" .

FROM gcr.io/distroless/static:nonroot
WORKDIR /qq
COPY --from=builder /app/bin/qq ./qq

ENTRYPOINT ["./qq"]
