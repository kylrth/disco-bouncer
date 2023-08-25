FROM golang:1.21 as builder

ARG TARGETARCH

COPY . /usr/src/disco-bouncer
WORKDIR /usr/src/disco-bouncer
ENV GOOS=linux
ENV GOARCH=$TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go get -v ./... && \
    CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o /bouncer ./cmd

FROM gcr.io/distroless/static-debian11
COPY --from=builder /bouncer /bouncer
USER 1000

ENTRYPOINT ["/bouncer"]
CMD ["serve"]
EXPOSE 80
