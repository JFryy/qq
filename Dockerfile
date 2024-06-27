FROM golang:1.22 as builder

WORKDIR /qq
COPY . .
ENV CGO_ENABLED 0
RUN go mod download && go mod verify
RUN apt update -y && apt install jq -y && make publish

FROM gcr.io/distroless/static:debug

COPY --from=builder /qq/bin/qq /
ENTRYPOINT ["/qq"]
