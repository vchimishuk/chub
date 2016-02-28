package player

import "github.com/vchimishuk/chub/vfs"

type PlaylistInfo struct {
	name     string
	duration int
	len      int
}

func (pi *PlaylistInfo) Name() string {
	return pi.name
}

func (pi *PlaylistInfo) SetName(name string) {
	pi.name = name
}

func (pi *PlaylistInfo) Duration() int {
	return pi.duration
}

func (pi *PlaylistInfo) Len() int {
	return pi.len
}

type Playlist struct {
	PlaylistInfo
	tracks []*vfs.Track
}

func NewPlaylist(name string) *Playlist {
	return &Playlist{
		PlaylistInfo: PlaylistInfo{name: name},
		tracks:       make([]*vfs.Track, 0)}
}

func (pl *Playlist) Tracks() []*vfs.Track {
	return pl.tracks
}

func (pl *Playlist) Len() int {
	return len(pl.tracks)
}

func (pl *Playlist) Clear() {
	pl.tracks = make([]*vfs.Track, 0)
	pl.duration = 0
}

func (pl *Playlist) Find(track *vfs.Track) int {
	for i := 0; i < len(pl.tracks); i++ {
		if &pl.tracks[i].Path == &track.Path {
			return i
		}
	}

	return -1
}

func (pl *Playlist) Append(tracks ...*vfs.Track) {
	pl.tracks = append(pl.tracks, tracks...)
	for _, t := range tracks {
		pl.duration += t.Length
	}
}

func (pl *Playlist) clone() *Playlist {
	tracks := make([]*vfs.Track, pl.Len())
	copy(tracks, pl.tracks)

	return &Playlist{
		PlaylistInfo: PlaylistInfo{
			name:     pl.name,
			duration: pl.duration},
		tracks: tracks}
}

func (pl *Playlist) info() *PlaylistInfo {
	return &PlaylistInfo{name: pl.name, duration: pl.duration, len: pl.len}
}
