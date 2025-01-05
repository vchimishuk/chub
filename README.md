**Chub** is a MPD-like audio player with rich CUE support.
Chub is inspired by the [MPD](https://www.musicpd.org) and
[MOC](http://moc.daper.net) audio players and incorporates key features from both: remote control and navigating an audio tracks collection as if it were a simple filesystem. Chub is implemented as a daemon application that is remotely controlled via separate client programs. Key features of Chub include track navigation and playback directly from the filesystem, as well as rich CUE integration, which allows album FLAC files to function as a set of separate audio tracks.

### Features
* Virtual FileSystem. Chub does not use database and does not require prior filesystem indexing. Tracks can be played and navigated directly from the filesystem in the same fashion as it is done by [MOC](http://moc.daper.net) player.
* Rich CUE support. Chub not only displays files from the filesystem but also represents them as audio tracks. For example, a single FLAC file can be displayed as multiple tracks from the same album.
* Remote control. Chub can be controlled using client applications running on the same or remote host. See section [Clients](#Clients) section for the list of available client applications.
* Playlists support.
* Rich audio files support. Chub uses FFMpeg to decode audio data, as a result it supports wide range of audio file formats as FFMpeg does.

### Clients
* [asp](https://github.com/vchimishuk/asp) -- ncurses client
* [chubc](https://github.com/vchimishuk/chubc) -- non-interactive terminal client
* [chubby](https://github.com/vchimishuk/chubby) -- Go client library

### OS and audio drivers support
* Supported output drivers: ALSA, OSS, ~~sndio~~.
* Supported OSes: GNU/Linux, FreeBSD, ~~OpenBSD~~.

### Build and run
The app can build and run using standard `go` command.
```
$ go build
$ go run main.go
```
It is also possible to easy build a package for some operation systems. See `dist` folder in the current source distribution.

### Configuration
Chub is configured using `~/.config/chub/chub.conf` configuration file. See man page and `chub.conf.example` file in the current source distribution for details.
