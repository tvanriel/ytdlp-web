#!/bin/env bash

set -eo pipefail

printf "%s\n" "[0/4] Downloading SHA256 checksums from GitHub"
curl --silent -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/SHA2-256SUMS.sig -o /tmp/SHA2-256SUMS.sig
curl --silent -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/SHA2-256SUMS -o /tmp/SHA2-256SUMS

printf "%s\n" "[1/4] Verifying GPG Signature of checksums  "
gpg --verify /tmp/SHA2-256SUMS.sig /tmp/SHA2-256SUMS

printf "%s\n" "[2/4] Downloading binary                  "
curl --silent -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /tmp/yt-dlp

printf "%s\n" "[3/4] Verifying checksum"
(cd /tmp; sha256sum -c --ignore-missing /tmp/SHA2-256SUMS)

printf "%s\n" "[4/4] Move to target    "
chmod +x /tmp/yt-dlp
mv "/tmp/yt-dlp" "${TARGET}/yt-dlp"
