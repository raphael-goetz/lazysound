package api

type Track struct {
	// Track identifier
	URN string `json:"urn"`
	// Track title
	Title string `json:"title"`
	// Track desription
	Description string `json:"description"`
	// Tempo
	Bpm int `json:"bpm"`
	// Duration
	Duration int64 `json:"duration"`
	/// User / Uploader of track
	User User `json:"user"`
}

type Tracks struct {
	// Collection of curret tracks cursor
	Collection []Track `json:"collection"`
	// Href of next tracks cursor
	Next string `json:"next_href"`
}
