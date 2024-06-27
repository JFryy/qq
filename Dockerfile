FROM golang:1.22 as builder

WORKDIR /qq
COPY . .
ENV CGO_ENABLED 0
RUN make build
RUN apt update -y && apt install jq -y && make test

FROM gcr.io/distroless/static:debug

COPY --from=builder /qq/bin/qq /
ENTRYPOINT ["/qq"]
