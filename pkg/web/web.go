package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/tvanriel/cloudsdk/http"
	"github.com/tvanriel/ytdlweb/assets"
	"github.com/tvanriel/ytdlweb/pkg/downloader"
	"github.com/tvanriel/ytdlweb/views"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Bucket string `hcl:"bucket"`
}

type WebOpts struct {
	fx.In
	Downloader *downloader.Downloader
	Logging    *zap.Logger
	S3         *minio.Client
	Config     Configuration
}

type Web struct {
	Downloader *downloader.Downloader
	Logging    *zap.Logger
	S3         *minio.Client
	Config     Configuration
}

// ApiGroup implements http.RouteGroup.
func (w *Web) ApiGroup() string {
	return ""
}

// Handler implements http.RouteGroup.
func (w *Web) Handler(e *echo.Group) {
	e.GET("bootstrap.css", w.bootstrapCSS)
	e.POST("api/add-video", w.Download)

	e.GET("api/thumbnail/:uuid", w.GetThumbnail)
	e.GET("api/audio/:uuid", w.GetAudio)
	e.GET("api/video/:uuid", w.GetVideo)
	e.GET("api/info/:uuid", w.ShowInfo)
	e.GET("api/delete/:uuid", w.Delete)
	e.GET("", w.indexhtml)
}

func (w *Web) Download(ctx echo.Context) error {
	form, err := ctx.FormParams()
	if err != nil {
		ctx.Response().WriteHeader(400)
		_, err := fmt.Fprintf(ctx.Response().Writer, "error: %s", err.Error())
		return err
	}

	videoId := form.Get("video_id")

	uuid, err := w.Downloader.Download(videoId)
	if err != nil {
		ctx.Response().WriteHeader(500)
		_, err := fmt.Fprintf(ctx.Response().Writer, "error: %s", err.Error())
		return err
	}

	return ctx.Redirect(301, "/?success="+uuid)
}

func (w *Web) bootstrapCSS(ctx echo.Context) error {
	ctx.Response().Header().Add("Content-Type", "text/css")
	ctx.Response().Writer.Write(assets.Bootstrap)
	return nil
}

func (w *Web) indexhtml(ctx echo.Context) error {
	objects := w.S3.ListObjects(ctx.Request().Context(), w.Config.Bucket, minio.ListObjectsOptions{
		Prefix:    "meta/",
		Recursive: true,
	})

	items := []views.RecentlyDownloadedEntry{}
	for o := range objects {
		if o.Err != nil {
			w.Logging.Error("object error", zap.Error(o.Err))
			continue
		}

		w.Logging.Info("object", zap.String("key", o.Key))

		entry, err := w.getMetadata(ctx.Request().Context(), o.Key)
		if err != nil {
			w.Logging.Error("cant fetch object from minio", zap.String("object", o.Key), zap.Error(err))
			continue
		}

		items = append(items, views.RecentlyDownloadedEntry{
			UUID:  entry.UUID,
			Title: entry.Title,
		})
	}
	w.Logging.Info("requested objects", zap.Int("len_items", len(items)))

	return views.Index(items, ctx.QueryParam("success") != "").Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (w *Web) metaKey(uuid string) string {
	return strings.Join([]string{"meta/", uuid, ".json"}, "")
}

func (w *Web) mediaKey(uuid, filename string) string {
	return strings.Join([]string{"media/", uuid, "/", filename}, "")
}

func (w *Web) getMetadata(ctx context.Context, key string) (*MetaEntry, error) {
	object, err := w.S3.GetObject(ctx, w.Config.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		w.Logging.Error("cant fetch meta object from minio", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	defer object.Close()

	d := json.NewDecoder(object)
	entry := MetaEntry{}
	err = d.Decode(&entry)
	if err != nil {
		w.Logging.Error("cannot decode json meta", zap.String("key", key), zap.Error(err))
		return nil, err

	}

	return &entry, err
}

func (w *Web) GetVideo(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	metaKey := w.metaKey(uuid)

	entry, err := w.getMetadata(ctx.Request().Context(), metaKey)
	if err != nil {
		w.Logging.Error("cannot get metadata for media", zap.String("uuid", uuid), zap.Error(err))
		return err
	}

	return w.getMedia(entry.Files.Video, uuid, ctx)
}

func (w *Web) GetAudio(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	metaKey := w.metaKey(uuid)
	entry, err := w.getMetadata(ctx.Request().Context(), metaKey)
	if err != nil {
		w.Logging.Error("cannot get metadata for media", zap.String("uuid", uuid), zap.Error(err))
		return err
	}
	return w.getMedia(entry.Files.Audio, uuid, ctx)
}

func (w *Web) GetThumbnail(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	metaKey := w.metaKey(uuid)
	entry, err := w.getMetadata(ctx.Request().Context(), metaKey)
	if err != nil {
		w.Logging.Error("cannot get metadata for media", zap.String("uuid", uuid), zap.Error(err))
		return err
	}
	return w.getMedia(entry.Files.Thumbnail, uuid, ctx)
}

func (w *Web) getMedia(filename, uuid string, ctx echo.Context) error {
	mediaKey := w.mediaKey(uuid, filename)
	w.Logging.Info("got media key", zap.String("key", mediaKey))

	o, err := w.S3.GetObject(ctx.Request().Context(), w.Config.Bucket, mediaKey, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	defer o.Close()

	contentDisposition := "inline"
	if ctx.QueryParam("download") == "1" {
		contentDisposition = "attachment"
	}
	ctx.Response().Writer.Header().Add("Content-Disposition", contentDisposition+"; filename="+strconv.Quote(filename))

	io.Copy(ctx.Response().Writer, o)

	return nil
}

func (w *Web) ShowInfo(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	metaKey := w.metaKey(uuid)
	entry, err := w.getMetadata(ctx.Request().Context(), metaKey)
	if err != nil {
		w.Logging.Error("cannot get metadata for media", zap.String("uuid", uuid), zap.Error(err))
		return err
	}

	return views.Info(views.VideoInfo{
		UUID:         entry.UUID,
		Title:        entry.Title,
		Description:  entry.Description,
		Channel:      entry.Channel,
		DurationText: entry.DurationText,
		Video:        entry.Files.Video,
		Audio:        entry.Files.Audio,
		Thumbnail:    entry.Files.Thumbnail,
		OriginalURL:  entry.OriginalURL,
		UploadDate:   time.Unix(int64(entry.Timestamp), 0).In(time.UTC).Format(time.DateTime),
	}).Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (w *Web) Delete(ctx echo.Context) error {
	uuid := ctx.Param("uuid")
	metaKey := w.metaKey(uuid)
	entry, err := w.getMetadata(ctx.Request().Context(), metaKey)
	if err != nil {
		return err
	}

	video := w.mediaKey(uuid, entry.Files.Video)
	audio := w.mediaKey(uuid, entry.Files.Audio)
	thumbnail := w.mediaKey(uuid, entry.Files.Thumbnail)

	err = errors.Join(
		w.S3.RemoveObject(ctx.Request().Context(), w.Config.Bucket, metaKey, minio.RemoveObjectOptions{
			ForceDelete: true,
		}),
		w.S3.RemoveObject(ctx.Request().Context(), w.Config.Bucket, video, minio.RemoveObjectOptions{
			ForceDelete: true,
		}),
		w.S3.RemoveObject(ctx.Request().Context(), w.Config.Bucket, audio, minio.RemoveObjectOptions{
			ForceDelete: true,
		}),
		w.S3.RemoveObject(ctx.Request().Context(), w.Config.Bucket, thumbnail, minio.RemoveObjectOptions{
			ForceDelete: true,
		}),
	)
	if err != nil {
		return err
	}
	return ctx.Redirect(307, "/")
}

// Version implements http.RouteGroup.
func (w *Web) Version() string {
	return ""
}

var _ http.RouteGroup = &Web{}

func NewWeb(o WebOpts) *Web {
	return &Web{
		Downloader: o.Downloader,
		Logging:    o.Logging.Named("web"),
		S3:         o.S3,
		Config:     o.Config,
	}
}
