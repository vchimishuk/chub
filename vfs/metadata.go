package vfs

import (
	"time"

	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/vfs/db"
)

type metadata struct {
	modified time.Time
	Artist   string
	Album    string
	Year     int
	Title    string
	Number   int
	Length   int
}

func getMetadata(path *Path) (*metadata, error) {
	// TODO: Do not fetch metadata for non-audio files.
	//       Skipp them in path.go?
	b, err := db.Get(path.File())
	if err != nil {
		return nil, err
	}

	fi, err := path.FileInfo()
	if err != nil {
		return nil, err
	}
	modtime := fi.ModTime()

	if len(b) != 0 {
		md := deserializeMetadata(b)
		if md.modified.Equal(modtime) {
			return md, nil
		}
	}

	fmd, err := format.GetMetadata(path.File())
	if err != nil {
		return nil, err
	}

	md := &metadata{
		modified: modtime,
		Artist:   fmd.Artist(),
		Album:    fmd.Album(),
		Year:     fmd.Year(),
		Title:    fmd.Title(),
		Number:   fmd.Number(),
		Length:   fmd.Length(),
	}

	err = db.Put(path.File(), serializeMetadata(md))
	if err != nil {
		return nil, err
	}

	return md, nil
}

func deserializeMetadata(buf []byte) *metadata {
	o := 0

	mod := bytesToInt64(buf[o : o+8])
	o += 8

	nar := int(buf[o])
	o++
	ar := string(buf[o : o+nar])
	o += nar

	nal := int(buf[o])
	o++
	al := string(buf[o : o+nal])
	o += nal

	yr := bytesToInt32(buf[o : o+4])
	o += 4

	ntl := int(buf[o])
	o++
	tl := string(buf[o : o+ntl])
	o += ntl

	nm := bytesToInt32(buf[o : o+4])
	o += 4

	ln := bytesToInt32(buf[o : o+4])
	o += 4

	return &metadata{
		modified: time.Unix(mod, 0),
		Artist:   ar,
		Album:    al,
		Year:     int(yr),
		Title:    tl,
		Number:   int(nm),
		Length:   int(ln),
	}
}

func serializeMetadata(md *metadata) []byte {
	var b []byte

	b = append(b, int64ToBytes(md.modified.Unix())...)
	b = append(b, byte(len(md.Artist)))
	b = append(b, []byte(md.Artist)...)
	b = append(b, byte(len(md.Album)))
	b = append(b, []byte(md.Album)...)
	b = append(b, int32ToBytes(int32(md.Year))...)
	b = append(b, byte(len(md.Title)))
	b = append(b, []byte(md.Title)...)
	b = append(b, int32ToBytes(int32(md.Number))...)
	b = append(b, int32ToBytes(int32(md.Length))...)

	return b
}

func bytesToInt32(b []byte) int32 {
	if len(b) != 4 {
		panic("four-element slice expected")
	}
	i := int32(0)
	i |= int32(b[3])
	i = i << 8
	i |= int32(b[2])
	i = i << 8
	i |= int32(b[1])
	i = i << 8
	i |= int32(b[0])

	return i
}

func int32ToBytes(i int32) []byte {
	return []byte{
		byte(i),
		byte(i >> 8),
		byte(i >> 16),
		byte(i >> 24),
	}
}

func bytesToInt64(b []byte) int64 {
	if len(b) != 8 {
		panic("four-element slice expected")
	}
	i := int64(0)
	i |= int64(b[7])
	i = i << 8
	i |= int64(b[6])
	i = i << 8
	i |= int64(b[5])
	i = i << 8
	i |= int64(b[4])
	i = i << 8
	i |= int64(b[3])
	i = i << 8
	i |= int64(b[2])
	i = i << 8
	i |= int64(b[1])
	i = i << 8
	i |= int64(b[0])

	return i
}

func int64ToBytes(i int64) []byte {
	return []byte{
		byte(i),
		byte(i >> 8),
		byte(i >> 16),
		byte(i >> 24),
		byte(i >> 32),
		byte(i >> 40),
		byte(i >> 48),
		byte(i >> 56),
	}
}
