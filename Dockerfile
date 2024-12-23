FROM --platform=$BUILDPLATFORM golang:latest as builder

ADD . /usr/src/ytdlp-web
WORKDIR /usr/src/ytdlp-web
COPY go.mod go.sum ./
RUN go mod download

ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /usr/bin/ytdlp-web .

FROM debian:stable-slim as final
COPY --from=builder /usr/bin/ytdlp-web /usr/bin/ytdlp-web

EXPOSE 8080

CMD ["/usr/bin/ytdlp-web"]
