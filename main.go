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

package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/vchimishuk/chub/alsa"
	"github.com/vchimishuk/chub/config"
	"github.com/vchimishuk/chub/format"
	"github.com/vchimishuk/chub/format/ffmpeg"
	"github.com/vchimishuk/chub/player"
	"github.com/vchimishuk/chub/server/cmd"
	"github.com/vchimishuk/chub/server/notif"
	"github.com/vchimishuk/chub/vfs"
	"github.com/vchimishuk/opt"
)

const (
	ProgName    = "chub"
	ProgVersion = "0.0.1"
)

var DefaultConfigFiles = []string{
	"~/.config/chub/chub.conf",
	"~/.chub.conf",
	"/etc/chub.conf",
}

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
	fmt.Println("Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>")
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

func readConfig(files []string) (*config.Config, error) {
	for _, fname := range files {
		fname, err := expandPath(fname)
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(fname)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		cfg, err := config.ParseFile(fname)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}

	return nil, nil
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

	var cfg *config.Config
	if name, ok := opts.String("config"); ok {
		c, err := readConfig([]string{name})
		if err != nil {
			fatal("failed to read configuration file: %s", err)
		} else if cfg == nil {
			fatal("file not found: %s", name)
		}
		cfg = c
	} else {
		c, err := readConfig(DefaultConfigFiles)
		if err != nil {
			fatal("failed to read configuration file: %s", err)
		} else if c == nil {
			c = &config.Config{}
		}
		cfg = c
	}

	ffmpegFmt := ffmpeg.NewFormat()

	// Initialize VFS.
	// vfs.RegisterFormat(flac.Format)
	// vfs.RegisterFormat(mp3.Format)
	format.Register(ffmpegFmt)

	err = vfs.SetRoot(cfg.String("vfs.root", "/"))
	if err != nil {
		panic(err)
	}

	// fmts := []player.Format{
	// 	flac.Format,
	// 	mp3.Format,
	// }
	output := alsa.New()
	pl := player.New([]format.Format{ffmpegFmt}, output)

	notifSrv := notif.NewServer(pl)
	notifSrv.Listen("127.0.0.1", 5225)
	fmt.Println("Notification server started")
	go notifSrv.Serve()

	cmdSrv := cmd.NewServer(pl)
	cmdSrv.Listen("127.0.0.1", 5115)
	fmt.Println("Command server started")
	cmdSrv.Serve()
	fmt.Println("Command server stopped")

	notifSrv.Close()
	fmt.Println("Notification server stopped")
}
