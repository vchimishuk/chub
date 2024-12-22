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

#include <fcntl.h>
#include <errno.h>
#include <unistd.h>
#include <sys/soundcard.h>
#include "oss.h"

int oss_open(const char *dev)
{
    return open(dev, O_WRONLY, 0);
}

int oss_close(int fd)
{
    return close(fd);
}

int oss_sample_rate(int fd, int rate)
{
    int r = rate;

    return ioctl(fd, SNDCTL_DSP_SPEED, &r);
}

int oss_channels(int fd, int channels)
{
    int c = channels;

    return ioctl(fd, SNDCTL_DSP_CHANNELS, &c);
}

int oss_format(int fd, int fmt)
{
    int f = fmt;

    return ioctl(fd, SNDCTL_DSP_SETFMT, &f);
}

int oss_write(int fd, const void *buf, int bufsz)
{
    return write(fd, buf, bufsz);
}

int oss_volume(int fd)
{
    int lvl;
    int e = ioctl(fd, SNDCTL_DSP_GETPLAYVOL, &lvl);
    if (e == -1) {
        return -1;
    }

    // Return right channel.
    return lvl >> 8;
}

int oss_setvolume(int fd, int vol)
{
    int lvl = (vol) | (vol << 8);

    return ioctl(fd, SNDCTL_DSP_SETPLAYVOL, &lvl);
}
