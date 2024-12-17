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

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	vconfig "github.com/vchimishuk/config"

	"github.com/vchimishuk/chub/alsa"
	"github.com/vchimishuk/chub/config"
	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/format/ffmpeg"
	"github.com/vchimishuk/chub/logger"
	"github.com/vchimishuk/chub/oss"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/server"
	"github.com/vchimishuk/chub/vfs"
	"github.com/vchimishuk/opt"
)

const (
	ProgName    = "chub"
	ProgVersion = "0.0.1"
)

var OptDescs = []*opt.Desc{
	{"c", "config", opt.ArgString, "FILE",
		"configuration file name"},
	{"h", "help", opt.ArgNone, "",
		"display this help and exit"},
	{"v", "version", opt.ArgNone, "",
		"output version information and exit"},
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: ", ProgName)
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func printUsage() {
	fmt.Printf("Usage: %s [OPTION]...\n", ProgName)
	fmt.Printf("Chub audio player daemon.\n")
	fmt.Printf("\n")
	fmt.Printf("The following OPTIONS are accepted:\n")
	fmt.Print(opt.Usage(OptDescs))
}

func printVersion() {
	fmt.Printf("%s %s\n", ProgName, ProgVersion)
	fmt.Println("Copyright 2016-2024 Viacheslav Chimishuk <vchimishuk@yandex.ru>")
	fmt.Println("Chub comes with ABSOLUTELY NO WARRANTY.")
	fmt.Println("You may redistribute copies of Chub")
	fmt.Println("under the terms of the GNU General Public License.")
	fmt.Println("For more information about these matters, see the file named COPYING.")
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		u, err := user.Current()
		if err != nil {
			return "", err
		}

		return filepath.Join(u.HomeDir, p[2:]), nil
	} else {
		return p, nil
	}
}

func parseConfig(file string) (*vconfig.Config, error) {
	if file != "" {
		f, err := expandPath(file)
		if err != nil {
			return nil, err
		}
		cfg, err := config.ParseFile(f)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("%s: %w", f, err)
		}

		return cfg, nil
	}

	f, err := expandPath("~/.config/chub/chub.conf")
	if err != nil {
		return nil, err
	}
	cfg, err := config.ParseFile(f)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: %w", f, err)
	}

	return cfg, nil
}

func main() {
	opts, args, err := opt.Parse(os.Args[1:], OptDescs)
	if err != nil {
		fatal(err.Error())
	}
	if len(args) != 0 {
		fatal("unexpected argument")
	}
	if opts.Bool("help") {
		printUsage()
		os.Exit(1)
	}
	if opts.Bool("version") {
		printVersion()
		os.Exit(1)
	}

	cfg, err := parseConfig(opts.StringOr("config", ""))
	if err != nil {
		fatal("%s", err)
	}
	if cfg == nil {
		cfg = &vconfig.Config{}
	}

	ffmpegFmt := ffmpeg.NewFormat()
	format.Register(ffmpegFmt)

	err = vfs.SetRoot(cfg.StringOr("vfs-root", "/"))
	if err != nil {
		panic(err)
	}

	var output player.Output
	switch cfg.StringOr("output", "alsa") {
	case "alsa":
		output = alsa.New()
	case "oss":
		output = oss.New()
	default:
		panic("unsupported output")
	}

	p := player.New([]format.Format{ffmpegFmt}, output)

	s := server.New(p)
	err = s.Listen(cfg.StringOr("server-host", "0.0.0.0"),
		cfg.IntOr("server-port", 5115))
	if err != nil {
		fatal("%s", err)
	}
	s.Serve()
	err = p.Close()
	if err != nil {
		logger.Error("failed to close player: %s", err)
	}
}
