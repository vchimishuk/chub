// Copyright 2016-2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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
	"errors"
	"slices"

	"github.com/vchimishuk/config"
)

var spec = &config.Spec{
	Strict: true,
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type:   config.TypeString,
			Name:   "output",
			Parser: parseEnum([]string{"alsa", "oss"}),
		},
		&config.PropertySpec{
			Type: config.TypeString,
			Name: "server-host",
		},
		&config.PropertySpec{
			Type: config.TypeInt,
			Name: "server-port",
		},
		&config.PropertySpec{
			Type: config.TypeString,
			Name: "vfs-root",
		},
	},
}

func ParseFile(path string) (*config.Config, error) {
	return config.ParseFile(spec, path)
}

// TODO: Use io.Reader instead of string.
func Parse(s string) (*config.Config, error) {
	return config.Parse(spec, s)
}

func parseEnum(vals []string) func(v any) (any, error) {
	return func(v any) (any, error) {
		s := v.(string)
		if !slices.Contains(vals, s) {
			return nil, errors.New("unsupported value")
		}

		return s, nil
	}
}
