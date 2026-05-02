package api

import "strconv"

type Track struct {
	// Track identifier
	ID int `json:"id"`
	// Track identifier
	URN string `json:"urn"`
	// Track title
	Title string `json:"title"`
	// Track desription
	Description string `json:"description"`
	// Artwork image
	ArtworkURL string `json:"artwork_url"`
	// Tag list
	TagList string `json:"tag_list"`
	// Waveform image
	WaveformURL string `json:"waveform_url"`
	// Tempo
	Bpm BPM `json:"bpm"`
	// Duration
	Duration int64 `json:"duration"`
	// Stream access level (e.g. playable, preview, blocked)
	Access string `json:"access"`
	/// User / Uploader of track
	User User `json:"user"`
}

type BPM float64

func (b *BPM) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || len(data) == 0 {
		*b = 0
		return nil
	}
	// handle quoted numbers
	if data[0] == '"' && data[len(data)-1] == '"' {
		s := string(data[1 : len(data)-1])
		if s == "" {
			*b = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			*b = 0
			return nil
		}
		*b = BPM(v)
		return nil
	}
	v, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		*b = 0
		return nil
	}
	*b = BPM(v)
	return nil
}

type Tracks struct {
	// Collection of curret tracks cursor
	Collection []Track `json:"collection"`
	// Href of next tracks cursor
	Next string `json:"next_href"`
}
