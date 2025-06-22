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

package config

import (
	"fmt"
	"os"
	"path"

	"github.com/vchimishuk/config"
)

var stateSpec = &config.Spec{
	Strict: true,
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type: config.TypeInt,
			Name: "volume",
		},
	},
}

type State struct {
	Volume int
}

func LoadState(path string) (*State, error) {
	c, err := config.ParseFile(stateSpec, path)
	if err != nil {
		return nil, err
	}

	return &State{
		Volume: c.IntOr("volume", 50),
	}, nil
}

func SaveState(p string, st *State) error {
	s := fmt.Sprintf("volume = %d\n", st.Volume)

	d, _ := path.Split(p)
	if d != "" {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(p, []byte(s), 0644)
}
