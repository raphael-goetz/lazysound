package api

type Playlist struct {
	/// Playlist title
	Title string `json:"title"`
	/// Playlist description
	Description string `json:"description"`
	/// Playlist identifier
	ID int `json:"id"`
	/// Playlist identifier
	URN string `json:"urn"`
	/// Duration
	Duration int64 `json:"duration"`
	/// Tracks inside the playlist
	Tracks []Track `json:"tracks"`
	/// Count of tracks inside the playlist
	TrackCount int64 `json:"track_count"`
	/// User / Creator of playlist
	User User `json:"user"`
	/// User / Creator identifier
	UserID int `json:"user_id"`
}

type Playlists struct {
	// List of playlists inside the current cursor
	Collection []Playlist `json:"collection"`
	// Href with the next cursor
	Next string `json:"next_href"`
}
