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

#ifndef FLAC_H
#define FLAC_H

#include <FLAC/all.h>

#define MAX_SUPPORTED_CHANNELS 2
#define SAMPLES_PER_WRITE 512
#define SAMPLE_BUF_SIZE ((FLAC__MAX_BLOCK_SIZE + SAMPLES_PER_WRITE) * \
        MAX_SUPPORTED_CHANNELS * (32 / 8))

#define SFMT_UNKNOWN 0
#define SFMT_S8 1
#define SFMT_S16_LE 2
#define SFMT_S32_LE 4

#define MIN(a, b) ((a) < (b) ? (a) : (b))

struct flac_decoder {
    FILE *file;
    FLAC__StreamDecoder *fsd;
    int abort;
    int len;
    unsigned int total_samples;
    FLAC__byte buf[SAMPLE_BUF_SIZE];
    unsigned int buf_fill;
    unsigned int channels;
    unsigned int format;
    unsigned int bits_per_sample;
    unsigned int sample_rate;
    int bitrate;
    int avg_bitrate;
    /* Last decoded sample number. */
    FLAC__uint64 last_decode_position;
    /* Total bytes decoded since beginning of the decoding process or */
    /* from last seek request if any. */
    FLAC__uint32 decoded_bytes;
    /* Second decoding started from. Initialy decoding starting from 0 second */
    /* but after seek call offset is set to seek seconds offset argument. */
    FLAC__uint32 time_offset;
    /* Current decoding time in seconds. */
    FLAC__uint32 time;
};

static size_t pack_pcm_signed(FLAC__byte *data,
    const FLAC__int32 * const input[], unsigned wide_samples,
    unsigned channels, unsigned bps);

struct flac_decoder *flac_open(const char *file);
void flac_close(struct flac_decoder *decoder);
int flac_decode(struct flac_decoder *decoder, char *buf, int len);
int flac_seek(struct flac_decoder *decoder, int pos, int rel);
int flac_time(struct flac_decoder *decoder);
int flac_sample_rate(struct flac_decoder *decoder);
int flac_channels(struct flac_decoder *decoder);
int flac_length(struct flac_decoder *decoder);

#endif // FLAC_H
