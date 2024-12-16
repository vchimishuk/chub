// Copyright 2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package assert

import (
	"testing"
)

func True(t *testing.T, a bool) {
	if !a {
		t.Fatal()
	}
}

func Nil(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func Error(t *testing.T, err error, s string) {
	if err == nil {
		t.Fatal("non-nil expected")
	}
	if err.Error() != s {
		t.Fatalf("`%s` != `%s`", err.Error(), s)
	}
}
