package config

import (
	"github.com/tvanriel/cloudsdk/hclconfig"
	"github.com/tvanriel/cloudsdk/http"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"github.com/tvanriel/cloudsdk/logging"
	"github.com/tvanriel/cloudsdk/s3"
	"github.com/tvanriel/ytdlweb/pkg/downloader"
	"github.com/tvanriel/ytdlweb/pkg/web"
	"go.uber.org/fx"
)

type Configuration struct {
	fx.Out

	Downloader       downloader.DownloaderConfig `hcl:"downloader,block"`
	HttpConfig       http.Configuration          `hcl:"http,block"`
	LogConfiguration logging.Configuration       `hcl:"logging,block"`
	Kubernetes       *kubernetes.Configuration   `hcl:"kubernetes,block"`
	S3Configuration  *s3.Configuration           `hcl:"s3,block"`
	WebConfiguration web.Configuration           `hcl:"web,block"`
}

func HclConfiguration() (Configuration, error) {
	c := Configuration{}

	err := hclconfig.HclConfiguration(&c, "ytdlweb")
	if err != nil {
		return Configuration{}, err
	}

	return c, nil
}
