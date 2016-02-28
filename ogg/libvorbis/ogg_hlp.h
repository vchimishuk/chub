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

#ifndef OGG_HLP_H
#defne OGG_HLP_H

#include <stdlib.h>
#include <vorbis/codec.h>
#include <vorbis/vorbisfile.h>

/*
 * Read size bytes from the stream into a buffer.
 * Returns actual number of read bytes.
 */
size_t ogg_hlp_read(OggVorbis_File *vf, char *buf, size_t size,
                    int bigendianp, int word, int sgned);

#endif // OGG_HLP_H
