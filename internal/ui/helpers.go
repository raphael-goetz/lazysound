package ui

import (
	"math/rand"

	api "github.com/raphael-goetz/lazysound/lib/soundcloud"
)

func playlistTrackIDs(tracks []api.Track) []int {
	ids := make([]int, 0, len(tracks))
	for _, t := range tracks {
		if t.ID > 0 {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

func containsInt(list []int, v int) bool {
	for _, it := range list {
		if it == v {
			return true
		}
	}
	return false
}

func removeInt(list []int, v int) []int {
	out := list[:0]
	for _, it := range list {
		if it != v {
			out = append(out, it)
		}
	}
	return out
}

func updatePlaylistSlice(list []api.Playlist, p api.Playlist) []api.Playlist {
	for i := range list {
		if list[i].ID == p.ID && p.ID != 0 {
			list[i] = p
			return list
		}
	}
	return list
}

func removePlaylistByID(list []api.Playlist, id int) []api.Playlist {
	out := list[:0]
	for _, p := range list {
		if p.ID != id {
			out = append(out, p)
		}
	}
	return out
}

func removeTrackByID(list []api.Track, id int) []api.Track {
	out := list[:0]
	for _, t := range list {
		if t.ID != id {
			out = append(out, t)
		}
	}
	return out
}

func likedTrackMap(list []api.Track) map[int]bool {
	out := make(map[int]bool)
	for _, t := range list {
		if t.ID != 0 {
			out[t.ID] = true
		}
	}
	return out
}

func likedPlaylistMap(list []api.Playlist) map[int]bool {
	out := make(map[int]bool)
	for _, p := range list {
		if p.ID != 0 {
			out[p.ID] = true
		}
	}
	return out
}

func updateTrackSlice(list []api.Track, t api.Track) []api.Track {
	for i := range list {
		if list[i].ID == t.ID && t.ID != 0 {
			list[i] = t
			return list
		}
	}
	return list
}

func updateTracksInPlaylists(playlists []api.Playlist, t api.Track) {
	for pi := range playlists {
		for ti := range playlists[pi].Tracks {
			if playlists[pi].Tracks[ti].ID == t.ID && t.ID != 0 {
				playlists[pi].Tracks[ti] = t
			}
		}
	}
}

func shuffleTracks(in []api.Track) []api.Track {
	out := make([]api.Track, len(in))
	copy(out, in)
	for i := len(out) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}
