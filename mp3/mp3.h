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

#ifndef MP3_H
#define MP3_H

#include <stdio.h>
#include <mad.h>

/* Size of the read buffer. */
#define BUFFER_SIZE (5 * 8192)

struct mp3_decoder {
    /* Sample rate of the stream. */
    int sample_rate;
    /* Number of channels in the stream. */
    int channels;
    /* Length of the file in seconds. */
    int length;
    /* Current decoding position in seconds. */
    int current_position;
    FILE *file;
    int current_sample;
    struct mad_stream stream;
    struct mad_frame frame;
    struct mad_header header;
    struct mad_synth synth;
    mad_timer_t timer;
    unsigned char buf[BUFFER_SIZE + MAD_BUFFER_GUARD];
};

struct mp3_decoder *mp3_open(const char *filename);
size_t mp3_decode(struct mp3_decoder *decoder, char *buf, size_t len);
void mp3_close(struct mp3_decoder *decoder);


#endif // MP3_H
