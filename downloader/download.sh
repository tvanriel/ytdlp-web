#!/bin/env bash

mkdir /tmp/download/media /tmp/download/meta -p || :


mc alias set target "$MINIO_URL" "$MINIO_ACCESS" "$MINIO_SECRET"

sponsorblock_categories="sponsor,preview,music_offtopic,selfpromo"
info_keys='{ "categories", "channel", "channel_follower_count", "channel_id", "channel_is_verified", "channel_url", "description", "duration", "duration_string", "extractor", "extractor_key", "fulltitle", "id", "original_url", "timestamp", "title", "upload_date", }'

if [ "$DEBUG" == true ]; then
    set -x
fi

(
    cd /tmp/download/media || exit 1;
    mkdir audio video thumbnail


    cd thumbnail || exit 1;
    # download thumbnail
    "${YT_DLP}" \
        "--write-thumbnail" \
        --write-thumbnail \
        --convert-thumbnails png \
        --restrict-filenames \
        -o "thumbnail:%(title)s.%(ext)s" \
        --skip-download \
        "${YOUTUBE_URL}" \
        --cookies "${COOKIES}"

    ls | jq -R '{thumbnail: . }' > ../../meta/.thumbnail.json

    mc cp -- * "target/${MINIO_BUCKET}/media/${MINIO_PREFIX}/"

    cd ../audio || exit 1;

    # download audio
    "${YT_DLP}" \
        "--sponsorblock-remove" "$sponsorblock_categories"\
        "-x" -f "ba*" \
        "--audio-format=vorbis" \
        "${YOUTUBE_URL}" \
        --restrict-filenames \
        -o "%(title)s.%(ext)s" \
        --cookies "${COOKIES}"

    ls | jq -R '{audio: .}' > ../../meta/.audio.json

    mc cp -- * "target/${MINIO_BUCKET}/media/${MINIO_PREFIX}/"

    cd ../video || exit 1;

    # download video
    "${YT_DLP}" \
        "${YOUTUBE_URL}" \
        "--sponsorblock-remove" "$sponsorblock_categories" \
        -f "bestvideo[height<=?1080]+bestaudio" \
        --restrict-filenames \
        -o "%(title)s.%(ext)s" \
        --cookies "${COOKIES}"

    mc cp -- * "target/${MINIO_BUCKET}/media/${MINIO_PREFIX}/"

    ls | jq -R '{video: .}' > ../../meta/.video.json

    cd ../

    mv video/* .
    mv audio/* .
    mv thumbnail/* .

    mediainfo --Output=JSON -- * | jq '{meta: .}' > ../meta/.meta.json

    cd ../meta || exit 1;

    (jq -sc '.[0] * .[1] * .[2]' .thumbnail.json .video.json .audio.json | jq '{files: .}' ) > .files.json


    # download info
    "${YT_DLP}" \
        "${YOUTUBE_URL}" \
        --no-download -j | \
        jq "$info_keys"> .info.json

    date --iso-8601=minutes | jq -R '{time: .}' > .time.json

    echo "$MINIO_PREFIX" | jq -R "{uuid: .}" > .prefix.json

    jq -sc '.[0] * .[1] * .[2] * .[3] * .[4]' .prefix.json .files.json .meta.json .time.json .info.json > metadata.json

    mc cp metadata.json "target/${MINIO_BUCKET}/meta/${MINIO_PREFIX}.json"
)


