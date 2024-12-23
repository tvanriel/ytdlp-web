package web

type MetaEntry struct {
	UUID         string `json:"uuid"`
	Title        string `json:"title"`
	Channel      string `json:"channel"`
	Description  string `json:"description"`
	DurationText string `json:"duration_string"`
	OriginalURL  string `json:"original_url"`
	Timestamp    int    `json:"timestamp"`
	Files        struct {
		Thumbnail string `json:"thumbnail"`
		Audio     string `json:"audio"`
		Video     string `json:"video"`
	} `json:"files"`
}
