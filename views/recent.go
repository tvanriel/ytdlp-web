package views

type RecentlyDownloadedEntry struct {
	UUID  string
	Title string
}

type VideoInfo struct {
	UUID         string
	Description  string
	Channel      string `json:"channel"`
	Title        string
	DurationText string

	Video, Audio, Thumbnail string

	OriginalURL string

	UploadDate string
}
