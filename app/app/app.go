package app

import (
	"github.com/tvanriel/cloudsdk/http"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"github.com/tvanriel/cloudsdk/logging"
	"github.com/tvanriel/cloudsdk/s3"
	"github.com/tvanriel/ytdlweb/pkg/config"
	"github.com/tvanriel/ytdlweb/pkg/downloader"
	"github.com/tvanriel/ytdlweb/pkg/web"
	"go.uber.org/fx"
)

func RunWeb() {
	app := fx.New(
		config.Module,
		web.Module,
		logging.Module,
		downloader.Module,
		kubernetes.Module,
		s3.Module,
		http.Module,
		logging.FXLogger(),
	)
	app.Run()
}
