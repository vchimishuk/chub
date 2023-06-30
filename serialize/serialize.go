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

package serialize

import (
	"bytes"
	"strconv"
)

type Serializable interface {
	Serialize() string
}

type mapSerializable map[string]any

func (m mapSerializable) Serialize() string {
	return Map(m)
}

func Wrap(m map[string]any) Serializable {
	return mapSerializable(m)
}

func Map(m map[string]any) string {
	var b bytes.Buffer
	var l int = len(m)
	var i int = 1

	for k, v := range m {
		b.WriteString(k)
		b.WriteString(": ")

		switch v.(type) {
		case int:
			b.WriteString(strconv.Itoa(v.(int)))
		case string:
			b.WriteString(strconv.Quote(v.(string)))
		case bool:
			b.WriteString(strconv.FormatBool(v.(bool)))
		default:
			panic("Unsupported type")
		}

		if i < l {
			b.WriteString(", ")
		}
		i++
	}

	return b.String()
}
