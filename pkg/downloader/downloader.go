package downloader

import (
	"maps"

	"github.com/google/uuid"
	"github.com/tvanriel/cloudsdk/kubernetes"
	"go.uber.org/fx"
	"go.uber.org/zap"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DownloaderConfig struct {
	SecretName   string            `hcl:"secret_name"`
	YTDLPBinary  string            `hcl:"ytdlp_binary"`
	Labels       map[string]string `hcl:"labels"`
	NodeSelector map[string]string `hcl:"node_selector"`
	Tolerations  []Toleration      `hcl:"toleration,block"`
}

type Toleration struct {
	Key      string `hcl:"key,optional"`
	Operator string `hcl:"operator,optional"`
	Value    string `hcl:"value,optional"`
	Effect   string `hcl:"effect,optional"`
}

func (t Toleration) AsToleration() v1.Toleration {
	return v1.Toleration{
		Key:      t.Key,
		Operator: v1.TolerationOperator(t.Operator),
		Value:    t.Value,
		Effect:   v1.TaintEffect(t.Effect),
	}
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

func (d *Downloader) Tolerations() []v1.Toleration {
	tolerations := make([]v1.Toleration, 0, len(d.Config.Tolerations))
	for t := range d.Config.Tolerations {
		tolerations = append(tolerations, d.Config.Tolerations[t].AsToleration())
	}

	return tolerations
}

func (d *Downloader) Download(url, username string) (string, error) {
	u := uuid.New().String()
	job := d.downloaderJob(url, u, username)
	err := d.kclient.RunJob(job)
	return u, err
}

func (d *Downloader) downloaderJob(url, jobUUID, username string) *batchv1.Job {
	backoffLimit := int32(6)
	ttlSeconds := int32(3600 * 24 * 2)

	labels := maps.Clone(d.Config.Labels)

	if username != "" {
		labels["ytdlp.mitaka.nl/username"] = username
	}

	labels["ytdlp.mitaka.nl/download-id"] = jobUUID

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "yt-dlp-download-" + jobUUID,
			Labels: labels,
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
					NodeSelector: d.Config.NodeSelector,
					Tolerations:  d.Tolerations(),
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
									Name:  "USERNAME",
									Value: username,
								},
								{
									Name:  "MINIO_PREFIX",
									Value: jobUUID,
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
