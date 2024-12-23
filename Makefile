UPDATER_IMAGE = mitaka8/yt-dlp-updater
DOWNLOADER_IMAGE = mitaka8/yt-dlp-downloader
WEB_IMAGE = mitaka8/yt-dlp-web

.PHONY: updater
updater:
	cd updater; docker buildx build -t ${UPDATER_IMAGE} --push --platform linux/amd64,linux/arm64 .
.PHONY: downloader
downloader:
	cd downloader; docker buildx build -t ${DOWNLOADER_IMAGE} --push --platform linux/amd64,linux/arm64 .

.PHONY: web
web:
	docker buildx build -t ${WEB_IMAGE} --push --platform linux/amd64,linux/arm64 .
