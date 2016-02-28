package cmd

import "github.com/vchimishuk/chub/vfs"

func entryToMap(e interface{}) map[string]interface{} {
	switch e.(type) {
	case *vfs.Dir:
		return dirToMap(e.(*vfs.Dir))
	case *vfs.Track:
		return trackToMap(e.(*vfs.Track))
	default:
		panic("unsupported type")
	}
}

func dirToMap(d *vfs.Dir) map[string]interface{} {
	return map[string]interface{}{
		"type": "dir",
		"path": d.Path.Val(),
		"name": d.Name,
	}
}

func trackToMap(t *vfs.Track) map[string]interface{} {
	m := map[string]interface{}{
		"type":   "track",
		"path":   t.Path.Val(),
		"length": t.Length,
	}
	if t.Tag != nil {
		m["artist"] = t.Tag.Artist
		m["album"] = t.Tag.Album
		m["title"] = t.Tag.Title
		m["number"] = t.Tag.Number
	}

	return m
}
