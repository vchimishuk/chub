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

// The code is based on MOC (Music On Console) written by Damian Pietras,
// which in turn is based on libxmms-flac by Josh Coalson.

#include <sys/stat.h>
#include <stdio.h>
#include <string.h>
#include "flac.h"

static FLAC__StreamDecoderReadStatus read_callback(
    const FLAC__StreamDecoder *fsd, FLAC__byte buf[],
    size_t *size, void *data)
{
    FILE *file = ((struct flac_decoder *) data)->file;
    FLAC__StreamDecoderReadStatus status;

    *size = fread(buf, 1, *size, file);
    if (ferror(file) != 0) {
        status = FLAC__STREAM_DECODER_READ_STATUS_ABORT;
    } else if (feof(file) != 0) {
        status = FLAC__STREAM_DECODER_READ_STATUS_END_OF_STREAM;
    } else {
        status = FLAC__STREAM_DECODER_READ_STATUS_CONTINUE;
    }

    return status;
}

static FLAC__StreamDecoderSeekStatus seek_callback(
    const FLAC__StreamDecoder *fsd,
    FLAC__uint64 offset, void *data)
{
    FILE *file = ((struct flac_decoder *) data)->file;

    if (fseek(file, offset, SEEK_SET) == 0) {
        return FLAC__STREAM_DECODER_SEEK_STATUS_OK;
    } else {
        return FLAC__STREAM_DECODER_SEEK_STATUS_ERROR;
    }
}

static FLAC__StreamDecoderTellStatus tell_callback(
    const FLAC__StreamDecoder *fsd,
    FLAC__uint64 *offset, void *data)
{
    FILE *file = ((struct flac_decoder *) data)->file;

    *offset = ftell(file);
    if (*offset == -1) {
        return FLAC__STREAM_DECODER_TELL_STATUS_ERROR;
    } else {
        return FLAC__STREAM_DECODER_TELL_STATUS_OK;
    }
}

static FLAC__StreamDecoderLengthStatus length_callback(
    const FLAC__StreamDecoder *fsd,
    FLAC__uint64 *len, void *data)
{
    FILE *file = ((struct flac_decoder *) data)->file;
    struct stat stats;

    if (fstat(fileno(file), &stats) != 0) {
        return FLAC__STREAM_DECODER_LENGTH_STATUS_ERROR;
    } else {
        *len = stats.st_size;

        return FLAC__STREAM_DECODER_LENGTH_STATUS_OK;
    }
}

static FLAC__bool eof_callback(const FLAC__StreamDecoder *fsd, void *data)
{
    FILE *file = ((struct flac_decoder *) data)->file;

    return feof(file) == 0 ? 0 : 1;
}

static FLAC__StreamDecoderWriteStatus write_callback(
    const FLAC__StreamDecoder *fsd,
    const FLAC__Frame *frame,
    const FLAC__int32 *const buf[], void *data)
{
    struct flac_decoder *decoder = (struct flac_decoder *) data;
    const unsigned wide_samples = frame->header.blocksize;

    if (decoder->abort) {
        return FLAC__STREAM_DECODER_WRITE_STATUS_ABORT;
    }

    decoder->buf_fill = pack_pcm_signed(decoder->buf, buf, wide_samples,
        decoder->channels, decoder->bits_per_sample);

    return FLAC__STREAM_DECODER_WRITE_STATUS_CONTINUE;
}

static void metadata_callback(const FLAC__StreamDecoder *fsd,
    const FLAC__StreamMetadata *metadata, void *data)
{
    struct flac_decoder *decoder = (struct flac_decoder *) data;

    if (metadata->type == FLAC__METADATA_TYPE_STREAMINFO) {
        const FLAC__StreamMetadata_StreamInfo *si = &(metadata->data.stream_info);

        decoder->total_samples = (unsigned int) (si->total_samples & 0xffffffff);
        decoder->bits_per_sample = si->bits_per_sample;
        decoder->channels = si->channels;
        decoder->sample_rate = si->sample_rate;
        decoder->len = decoder->total_samples / decoder->sample_rate;

        switch (si->bits_per_sample / 8) {
        case 1:
            decoder->format = SFMT_S8;
            break;
        case 2:
            decoder->format = SFMT_S16_LE;
            break;
        case 3:
            decoder->format = SFMT_S32_LE;
            break;
        default:
            decoder->format = SFMT_UNKNOWN;
            break;
        }
    }
}

static void error_callback(const FLAC__StreamDecoder *fsd,
    FLAC__StreamDecoderErrorStatus status, void *data)
{
    struct flac_decoder *decoder = (struct flac_decoder *) data;

    if (status != FLAC__STREAM_DECODER_ERROR_STATUS_LOST_SYNC) {
        decoder->abort = 1;
    }
}

/* Convert FLAC big-endian data into PCM little-endian. */
static size_t pack_pcm_signed(FLAC__byte *data,
    const FLAC__int32 * const input[], unsigned wide_samples,
    unsigned channels, unsigned bps)
{
    FLAC__byte * const start = data;
    FLAC__int32 sample;
    const FLAC__int32 *input_;
    unsigned samples, channel;
    unsigned bytes_per_sample;
    unsigned incr;

    if (bps == 24) {
        bps = 32; /* we encode to 32-bit words */
    }
    bytes_per_sample = bps / 8;
    incr = bytes_per_sample * channels;

    for (channel = 0; channel < channels; channel++) {
        samples = wide_samples;
        data = start + bytes_per_sample * channel;
        input_ = input[channel];

        while(samples--) {
            sample = *input_++;

            switch(bps) {
            case 8:
                data[0] = sample;
                break;
            case 16:
                data[1] = (FLAC__byte)(sample >> 8);
                data[0] = (FLAC__byte)sample;
                break;
            case 32:
                data[3] = (FLAC__byte)(sample >> 16);
                data[2] = (FLAC__byte)(sample >> 8);
                data[1] = (FLAC__byte)sample;
                data[0] = 0;
                break;
            }

            data += incr;
        }
    }

    return wide_samples * channels * bytes_per_sample;
}

struct flac_decoder *flac_open(const char *file)
{
    struct flac_decoder *decoder = malloc(sizeof(struct flac_decoder));

    if (decoder == NULL) {
        return NULL;
    }

    decoder->file = NULL;
    decoder->fsd = NULL;
    decoder->abort = 0;
    decoder->len = -1;
    decoder->total_samples = 0;
    decoder->buf_fill = 0;
    decoder->channels = 0;
    decoder->format = SFMT_UNKNOWN;
    decoder->bits_per_sample = 0;
    decoder->sample_rate = 0;
    decoder->bitrate = 0;
    decoder->avg_bitrate = 0;
    decoder->last_decode_position = 0;
    decoder->decoded_bytes = 0;
    decoder->time_offset = 0;
    decoder->time = 0;

    decoder->file = fopen(file, "r");
    if (decoder->file == NULL) {
        flac_close(decoder);
        return NULL;
    }
    decoder->fsd = FLAC__stream_decoder_new();
    if (decoder->fsd == NULL) {
        flac_close(decoder);
        return NULL;
    }
    FLAC__stream_decoder_set_md5_checking(decoder->fsd, (FLAC__bool) 0);
    FLAC__stream_decoder_set_metadata_ignore_all(decoder->fsd);
    FLAC__stream_decoder_set_metadata_respond(decoder->fsd,
        FLAC__METADATA_TYPE_STREAMINFO);

    FLAC__StreamDecoderInitStatus st = FLAC__stream_decoder_init_stream(
        decoder->fsd,
        read_callback,
        seek_callback,
        tell_callback,
        length_callback,
        eof_callback,
        write_callback,
        metadata_callback,
        error_callback,
        decoder);
    if (st != FLAC__STREAM_DECODER_INIT_STATUS_OK) {
        flac_close(decoder);
        return NULL;
    }
    if (!FLAC__stream_decoder_process_until_end_of_metadata(decoder->fsd)) {
        flac_close(decoder);
        return NULL;
    }

    decoder->avg_bitrate = (decoder->bits_per_sample) * decoder->sample_rate;

    return decoder;
}

void flac_close(struct flac_decoder *decoder)
{
    if (decoder->fsd) {
        FLAC__stream_decoder_finish(decoder->fsd);
        FLAC__stream_decoder_delete(decoder->fsd);
    }
    if (decoder->file) {
        fclose(decoder->file);
    }
    free(decoder);
}

int flac_decode(struct flac_decoder *decoder, char *buf, int len)
{
    FLAC__StreamDecoder *fsd = decoder->fsd;
    int bytes_per_sample = decoder->bits_per_sample / 8;

    if (decoder->buf_fill == 0) {
        FLAC__uint64 decode_position;

        if (FLAC__stream_decoder_get_state(fsd) == FLAC__STREAM_DECODER_END_OF_STREAM) {
            return 0;
        }
        if (!FLAC__stream_decoder_process_single(fsd)) {
            return 0;
        }
        /* Count the bitrate */
        if(!FLAC__stream_decoder_get_decode_position(fsd, &decode_position)) {
            decode_position = 0;
        }

        int bytes_per_sec = bytes_per_sample * decoder->sample_rate
            * decoder->channels;

        if (decode_position > decoder->last_decode_position) {
            decoder->bitrate = (decode_position - decoder->last_decode_position) * 8.0
                / (decoder->buf_fill / (float) bytes_per_sec) / 1000;
        }

        decoder->last_decode_position = decode_position;
    }

    unsigned int to_copy = MIN((unsigned) len, decoder->buf_fill);
    memcpy(buf, decoder->buf, to_copy);
    memmove(decoder->buf, decoder->buf + to_copy, decoder->buf_fill - to_copy);
    decoder->buf_fill -= to_copy;

    decoder->decoded_bytes += to_copy;
    FLAC__uint32 decoded_samples = decoder->decoded_bytes / bytes_per_sample;
    FLAC__uint32 decoded_time = decoded_samples / decoder->sample_rate
        / decoder->channels;
    decoder->time = decoder->time_offset + decoded_time;

    return to_copy;
}

int flac_seek(struct flac_decoder *decoder, int pos, int rel)
{
    if (rel) {
        pos = decoder->time + pos;
    }

    if (pos > decoder->len || pos < 0) {
        return -1;
    }

    FLAC__uint64 pos_samples = decoder->sample_rate * pos;

    if (FLAC__stream_decoder_seek_absolute(decoder->fsd, pos_samples)) {
        decoder->decoded_bytes = 0;
        decoder->time_offset = pos;
        decoder->time = pos;

        return pos;
    } else {
        return -1;
    }
}

int flac_time(struct flac_decoder *decoder)
{
    return decoder->time;
}

int flac_sample_rate(struct flac_decoder *decoder)
{
    return decoder->sample_rate;
}
int flac_channels(struct flac_decoder *decoder)
{
    return decoder->channels;
}

int flac_length(struct flac_decoder *decoder)
{
    return decoder->len;
}
