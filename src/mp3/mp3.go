// mp3 package provides MP3 files support.
package mp3

import (
	"../vfs"
	"strings"
)

func match(file *vfs.Path) bool {
	ext := strings.ToLower(file.Ext())

	return ext == ".mp3"
}
