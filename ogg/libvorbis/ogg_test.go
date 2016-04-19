// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package libvorbis

import (
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
