package googlepickerjs

type ThumbnailObject struct {
	Height int    `javascript:"height"`
	Type   string `javascript:"type"`
	Width  int    `javascript:"width"`
	URL    string `javascript:"url"`
}
