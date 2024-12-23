FROM debian:latest

ARG TARGETPLATFORM
ARG TARGETARCH

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y ffmpeg curl jq mediainfo ca-certificates

RUN curl https://dl.min.io/client/mc/release/linux-$TARGETARCH/mc -o /usr/local/bin/mc

RUN chmod +x /usr/local/bin/mc

RUN addgroup --gid 1001 ytdlp
RUN adduser --gid 1001 --shell /bin/false --disabled-password --home /home/ytdlp --uid 1001 ytdlp
RUN mkdir -p /var/log/ytdlp
RUN chown ytdlp:ytdlp /var/log/ytdlp

COPY download.sh /usr/local/bin/download.sh

USER ytdlp

CMD ["download.sh"]
