FROM debian:buster AS builder

RUN apt-get update \
    && apt-get install -y \
        wget \
        gnupg \
        ca-certificates \
        git \
        jq \
        make \
    && rm -rf /var/lib/apt/lists/*

RUN wget -O /tmp/go.tar.gz https://golang.org/dl/go1.22.4.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOBIN="/go/bin"
WORKDIR /app
COPY . .
RUN make build


FROM gcr.io/distroless/static:nonroot
WORKDIR /qq
COPY --from=builder /app/bin/qq ./qq

ENTRYPOINT ["./qq"]
CMD ["./qq", "--help"]
