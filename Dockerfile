FROM golang:1.23rc1-bullseye AS builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED=1
RUN go build -o bin/qq -ldflags="-linkmode external -extldflags -static" .

FROM gcr.io/distroless/static:nonroot
WORKDIR /qq
COPY --from=builder /app/bin/qq ./qq

ENTRYPOINT ["./qq"]
