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

#ifndef MAD_HLP_H
#define MAD_HLP_H

#include <stdio.h>
#include <stdint.h>
#include <limits.h>
#include <string.h>
#include <mad.h>

/* Size of the read buffer. */
#define BUFFER_SIZE (5 * 8192)

/*
 * MAD decoder wrapper structure.
 */
struct gomad_decoder {
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


/*
 * Initialize MAD decoder.
 * Returns -1 on failure.
 */
int gomad_open(struct gomad_decoder *decoder, const char *filename);


/*
 * Read and decode len bytes form input file.
 */
size_t gomad_read(struct gomad_decoder *decoder, char *buf, size_t len);


/*
 * Close MAD decoder and free all related resources.
 */
void gomad_close(struct gomad_decoder *decoder);

#endif // MAD_HLP_H
