FROM debian:latest

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y curl gpg ca-certificates

COPY ./yt-dlp.key /tmp/yt-dlp.key

COPY ./update.sh /usr/local/bin/update

RUN addgroup --gid 1001 ytdlp
RUN adduser --gid 1001 --shell /bin/false --disabled-password --home /home/ytdlp --uid 1001 ytdlp
RUN mkdir -p /var/log/ytdlp
RUN chown ytdlp:ytdlp /var/log/ytdlp

USER ytdlp

ENV GNUPGHOME=/home/ytdlp/.gnupg HOME=/home/ytdlp

RUN cat /tmp/yt-dlp.key | gpg --import

CMD [ "update" ]
