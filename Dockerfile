FROM golang:1.22 as builder

WORKDIR /app
COPY ./ ./
RUN go mod tidy
COPY . .
ENV CGO_ENABLED 0
RUN make build

FROM gcr.io/distroless/static:debug

COPY --from=builder /app/bin/qq /
ENTRYPOINT ["/qq"]
CMD ["--help"]
