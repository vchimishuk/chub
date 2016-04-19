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

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <limits.h>
#include <string.h>
#include "mp3.h"

#define CLAMP(x, l, h) (((x) > (h)) ? (h) : (((x) < (l)) ? (l) : (x)))

/*
 * Fill buffer with new data from input file.
 * Returns size of actual data in the buffer.
 */
static size_t mp3_fill_buffer(struct mp3_decoder *decoder)
{
    size_t offset;
    size_t free_size;
    size_t read_len;

    if (decoder->stream.next_frame == NULL) {
        /* Place contents at the beginning */
        offset = 0;
    } else {
        /*
         * We still need to save the fragmented frame. Copy it
         * to the beginning and load as much as possible.
         */
        offset = decoder->stream.bufend - decoder->stream.next_frame;
        memmove(decoder->buf, decoder->stream.next_frame, offset);
    }

    free_size = BUFFER_SIZE - offset;
    read_len = fread(decoder->buf + offset, 1, free_size, decoder->file);

    if (read_len < free_size) {
        if (feof(decoder->file)) {
            if (read_len == 0) {
                return 0;
            }

            /*
             * Place some null bytes at the end
             * because we may overrun it
             */
            memset(decoder->buf + offset + read_len, 0x00000000, MAD_BUFFER_GUARD);
            read_len += MAD_BUFFER_GUARD;
        } else if (ferror(decoder->file)) {
            return -1;
        }
    }

    return offset + read_len;
}

/*
 * Decode frame.
 * Returns -1 on error.
 */
static int mp3_read_frame(struct mp3_decoder *decoder)
{
    size_t data_size;

    for (; ;) {
        if (decoder->stream.buffer == NULL || decoder->stream.error == MAD_ERROR_BUFLEN) {
            data_size = mp3_fill_buffer(decoder);
            if (data_size <= 0) {
                return -1;
            }

            mad_stream_buffer(&decoder->stream, decoder->buf, data_size);
            decoder->stream.error = MAD_ERROR_NONE;
        }

        if (mad_header_decode(&decoder->header, &decoder->stream) == 0) {
            mad_timer_add(&decoder->timer, decoder->frame.header.duration);
            decoder->position = decoder->timer.seconds;
            decoder->frame.header = decoder->header;
            return 0;
        }

        if (!MAD_RECOVERABLE(decoder->stream.error) && (decoder->stream.error != MAD_ERROR_BUFLEN)) {
            return -1;
        }
    }
}

/*
 * Reinitialize decoder, so file will be decoded from the beginning.
 */
static void mp3_rewind(struct mp3_decoder *decoder)
{
    rewind(decoder->file);

    decoder->position = 0;
    decoder->current_sample = 0;

    mad_stream_init(&decoder->stream);
    mad_frame_init(&decoder->frame);
    mad_header_init(&decoder->header);
    mad_synth_init(&decoder->synth);
    mad_timer_reset(&decoder->timer);
}

/*
 * Calculate and returns length of the file in seconds.
 */
static void mp3_fill_info(struct mp3_decoder *decoder)
{
    /*
     * There are three ways of calculating the length of an mp3:
     * 1) Constant bitrate: One frame can provide the information
     *    needed: # of frames and duration. Just see how long it
     *    is and do the division.
     * 2) Variable bitrate: Xing tag. It provides the number of
     *    frames. Each frame has the same number of samples, so
     *    just use that.
     * 3) All: Count up the frames and duration of each frame
     *    by decoding each one. We do this if we've no other
     *    choice, i.e. if it's a VBR file with no Xing tag.
     *
     * See source code of mocp project.
     */

    /*
     * TODO: This method can be improved and realized with algorythm
     *       described above. Now it is done in the same way
     *       as in herrie audio player.
     */

    off_t offset;
    off_t total;
    int first_frame = 1;

    do {
        if (mp3_read_frame(decoder) != 0) {
            decoder->length = decoder->timer.seconds;
            offset = ftello(decoder->file);
            mp3_rewind(decoder);

            return;
        }

        if (first_frame) {
            if (mad_frame_decode(&decoder->frame, &decoder->stream) != -1) {
                decoder->sample_rate = decoder->frame.header.samplerate;
                decoder->channels = MAD_NCHANNELS(&decoder->frame.header);
                first_frame = 0;
            }
        }
    } while (decoder->timer.seconds < 420);

    /* Extrapolate the time. Not really accurate, but good enough */
    offset = ftello(decoder->file);
    fseek(decoder->file, 0, SEEK_END);
    total = ftello(decoder->file);

    decoder->length = ((double) total / offset) * decoder->timer.seconds;

    mp3_rewind(decoder);

    return;
}

/*
 * Convert a fixed point sample to a short.
 */
static inline int16_t mp3_fixed_to_short(mad_fixed_t fixed)
{
    if (fixed >= MAD_F_ONE) {
        return SHRT_MAX;
    } else if (fixed <= -MAD_F_ONE) {
        return -SHRT_MAX;
    }
    return fixed >> (MAD_F_FRACBITS - 15);
}

static int fsize(FILE *file)
{
    int pos = ftell(file);
    int size;

    fseek(file, 0, SEEK_END);
    size = ftell(file);
    fseek(file, pos, SEEK_SET);

    return size;
}

struct mp3_decoder *mp3_open(const char *filename)
{
    struct mp3_decoder *d = malloc(sizeof(struct mp3_decoder));

    if (d == NULL) {
        return NULL;
    }

    d->file = fopen(filename, "r");
    if (d->file == NULL) {
        free(d);

        return NULL;
    }

    d->file_size = fsize(d->file);
    mp3_rewind(d);
    mp3_fill_info(d);

    return d;
}

size_t mp3_decode(struct mp3_decoder *decoder, char *buf, size_t len)
{
    int16_t *words_buf = (int16_t *)buf;
    size_t words_len = len / 2;

    size_t written = 0;
    int i;

    do {
        if (decoder->current_sample == 0) {
            if (mp3_read_frame(decoder) == -1) {
                return written * 2;
            }
            if (mad_frame_decode(&decoder->frame, &decoder->stream) == -1) {
                continue;
            }

            /* decoder->sample_rate = decoder->frame.header.samplerate; */
            /* decoder->channels = MAD_NCHANNELS(&decoder->frame.header); */

            mad_synth_frame(&decoder->synth, &decoder->frame);
        }

        while ((decoder->current_sample < decoder->synth.pcm.length) && (written < words_len)) {
            for (i = 0; i < MAD_NCHANNELS(&decoder->frame.header); i++) {
                words_buf[written++] = mp3_fixed_to_short(decoder->synth.pcm.samples[i][decoder->current_sample]);

            }

            decoder->current_sample++;
        }

        /* Move to the next frame. */
        if (decoder->current_sample == decoder->synth.pcm.length) {
            decoder->current_sample = 0;
        }
    } while (written < words_len);

    return written * 2;
}

void mp3_seek(struct mp3_decoder *decoder, int pos, int rel)
{
    if (rel) {
        pos += decoder->position;
    }

    /* Calculate the new relative position. */
    pos = CLAMP(pos, 0, (int) decoder->length);
    int new_pos = ((double) pos / decoder->length) * decoder->file_size;

    mp3_rewind(decoder);
    fseek(decoder->file, new_pos, SEEK_SET);
    decoder->timer.seconds = pos;
    decoder->timer.fraction = 0;
}

void mp3_close(struct mp3_decoder *decoder)
{
    mad_frame_finish(&decoder->frame);
    mad_stream_finish(&decoder->stream);
    mad_synth_finish(&decoder->synth);

    fclose(decoder->file);
    free(decoder);
}
