package daemon

import api "github.com/raphael-goetz/lazysound/lib/soundcloud"

type Request struct {
	Cmd      string      `json:"cmd"`
	Token    string      `json:"token,omitempty"`
	Track    *api.Track  `json:"track,omitempty"`
	Queue    []api.Track `json:"queue,omitempty"`
	StartIdx int         `json:"start_idx,omitempty"`
	Volume   int         `json:"volume,omitempty"`
	SeekSec  int         `json:"seek_sec,omitempty"`
	Shuffle  *bool       `json:"shuffle,omitempty"`
	Repeat   *bool       `json:"repeat,omitempty"`
}

type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	State *State `json:"state,omitempty"`
	Probe *Probe `json:"probe,omitempty"`
}

type Probe struct {
	TrackID int    `json:"track_id"`
	Badge   string `json:"badge"`
	Detail  string `json:"detail,omitempty"`
}

type State struct {
	Playing    bool       `json:"playing"`
	Paused     bool       `json:"paused"`
	Track      *api.Track `json:"track,omitempty"`
	Index      int        `json:"index"`
	Volume     int        `json:"volume"`
	Shuffle    bool       `json:"shuffle"`
	Repeat     bool       `json:"repeat"`
	ElapsedMs  int64      `json:"elapsed_ms,omitempty"`
	DurationMs int64      `json:"duration_ms,omitempty"`
	LastError  string     `json:"last_error,omitempty"`
}
