package downloader

import "go.uber.org/fx"

var Module = fx.Module("downloader",
	fx.Provide(NewDownloader),
)
