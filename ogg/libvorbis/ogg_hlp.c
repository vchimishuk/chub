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
// along with Chub.  If not, see <http://www.gnu.org/licenses/>.

#include "ogg_hlp.h"

size_t ogg_hlp_read(OggVorbis_File *vf, char *buf, size_t size,
    int bigendianp, int word, int sgned)
{
    int read = 0; /* Already read bytes. */

    while (read < size) {
        long rr = ov_read(vf, buf + read, size - read,
            bigendianp, word, sgned, NULL);
        if (rr <= 0) {
            break;
        }

        read += rr;
    }

    return read;
}
