logging {
  development = true
}

http {
  address = "0.0.0.0:8080"
  debug = true
  rate_limit = 20
}

web {
  bucket = "ytdlp"
}

kubernetes {
  in_cluster = false
  kubeconfig = "/path/to/kubeconfig"
  namespace = "ytdlp-web"
  api_server = "kubernetes:6443"
}

s3 {
  endpoint = ""
  access_key = ""
  secret_key = ""
  ssl = true
}

downloader {
	secret_name = "downloader-env"
	labels     = {
    "app.kubernetes.io/component": "downloader",
  }
  ytdlp_binary="yt-dlp"
}
