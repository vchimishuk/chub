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

#ifndef CHUB_OSS_H
#define CHUB_OSS_H

int oss_open(const char *dev);
int oss_close(int fd);
int oss_sample_rate(int fd, int rate);
int oss_channels(int fd, int channels);
int oss_format(int fd, int fmt);
int oss_write(int fd, const void *buf, int bufsz);
int oss_volume(int fd);
int oss_setvolume(int fd, int vol);

#endif // CHUB_OSS_H
