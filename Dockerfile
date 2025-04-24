FROM golang:1.24 AS builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED=1
RUN go build -o bin/qq -ldflags="-linkmode external -extldflags -static" .

FROM gcr.io/distroless/static:nonroot
WORKDIR /qq
COPY --from=builder /app/bin/qq ./qq

ENTRYPOINT ["./qq"]
