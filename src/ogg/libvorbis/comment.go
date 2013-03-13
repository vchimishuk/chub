package libvorbis

import "C"

// Popular user comment names.
const (
	CommentArtist = "ARTIST"
	CommentAlbum = "ALBUM"
	CommentTitle = "TITLE"
	CommentTrackNumber = "TRACKNUMBER"
)

// The Comment structure defines an Ogg Vorbis comment.
type Comment struct {
	UserComments []string
	Vendor       string
}
