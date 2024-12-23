package downloader

import (
	"github.com/google/uuid"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"go.uber.org/fx"
	"go.uber.org/zap"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DownloaderConfig struct {
	SecretName  string            `hcl:"secret_name"`
	YTDLPBinary string            `hcl:"ytdlp_binary"`
	Labels      map[string]string `hcl:"labels"`
}
type Downloader struct {
	kclient *kubernetes.KubernetesClient
	logger  *zap.Logger
	Config  DownloaderConfig
}

type NewDownloaderOpts struct {
	fx.In

	Config  DownloaderConfig
	Kclient *kubernetes.KubernetesClient
	Logger  *zap.Logger
}

func NewDownloader(o NewDownloaderOpts) *Downloader {
	return &Downloader{
		kclient: o.Kclient,
		logger:  o.Logger.Named("downloader"),
		Config:  o.Config,
	}
}

func (d *Downloader) Download(url string) (string, error) {
	u := uuid.New().String()
	job := d.downloaderJob(url, u)
	err := d.kclient.RunJob(job)
	return u, err
}

func (d *Downloader) downloaderJob(url, u string) *batchv1.Job {
	backoffLimit := int32(6)
	ttlSeconds := int32(3600 * 24 * 2)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "yt-dlp-download-" + u,
			Labels: d.Config.Labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSeconds,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: d.Config.Labels,
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: "binary",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/opt/k3s/yt-dlp",
								},
							},
						},
						{
							Name: "cookies",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "cookies",
									},
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": "metroid",
					},
					Tolerations: []v1.Toleration{
						{
							Key:      "cloud",
							Value:    "false",
							Operator: v1.TolerationOpEqual,
							Effect:   v1.TaintEffectNoExecute,
						},
					},
					Containers: []v1.Container{
						{
							Name:            "downloader",
							Image:           "mitaka8/yt-dlp-downloader:latest",
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "binary",
									ReadOnly:  true,
									MountPath: "/opt/yt-dlp",
								},
								{
									Name:      "cookies",
									MountPath: "/home/ytdlp/cookies",
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "YOUTUBE_URL",
									Value: url,
								},
								{
									Name:  "COOKIES",
									Value: "/home/ytdlp/cookies/cookies.txt",
								},
								{
									Name:  "YT_DLP",
									Value: "/opt/yt-dlp/yt-dlp",
								},
								{
									Name:  "MINIO_PREFIX",
									Value: u,
								},
							},
							EnvFrom: []v1.EnvFromSource{{
								SecretRef: &v1.SecretEnvSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: d.Config.SecretName,
									},
								},
							}},
						},
					},
				},
			},
		},
	}
}
