package libvorbis

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	// const filename = "test.ogg"

	// file, err := New(filename)
	// if err != nil {
	// 	t.Fatalf("New() failed. %s", err)
	// }
	// defer file.Close()

	// fmt.Printf("Filename: %s\n", filename)

	// // TotalTime
	// fmt.Printf("TimeTotal: %f\n", file.TimeTotal())

	// // Comment
	// comment := file.Comment()
	// fmt.Printf("Comment:\n")
	// fmt.Printf("\tVendor: %s\n", comment.Vendor)
	// fmt.Printf("\tUser comments:\n")
	// for _, uc := range comment.UserComments {
	// 	fmt.Printf("\t\t%s\n", uc)
	// }

	// // Info
	// info := file.Info()
	// fmt.Printf("Info:\n")
	// fmt.Printf("\tVersion: %d\n", info.Version)
	// fmt.Printf("\tChannels: %d\n", info.Channels)
	// fmt.Printf("\tRate: %d\n", info.Rate)
	// fmt.Printf("\tBitrateUpper: %d\n", info.BitrateUpper)
	// fmt.Printf("\tBitrateNominal: %d\n", info.BitrateNominal)
	// fmt.Printf("\tBitrateLower: %d\n", info.BitrateLower)
	// fmt.Printf("\tBitrateWindow: %d\n", info.BitrateWindow)

	// // Read
	// fmt.Printf("Decoding...    ")

	// secPerPercent := file.TimeTotal() / 100
	// buf := make([]byte, 4096)
	// for file.Read(buf) > 0 {
	// 	percentsDone := uint(file.TimeTell() / secPerPercent)
	// 	fmt.Printf("%d%%\n", percentsDone)
	// }
}
