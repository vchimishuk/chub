// Copyright 2019 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

#include <sys/param.h>
#include <libavutil/opt.h>
#include <libavutil/rational.h>
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libswresample/swresample.h>
#include "ffmpeg.h"


static void *zmalloc(size_t n)
{
    void *p = malloc(n);
    memset(p, 0, n);

    return p;
}

// Send next packet for decoding.
// Returns 0 if success or negative number in case of error.
static int ffmpeg_send_packet(struct ffmpeg_file *file)
{
    for (;;) {
        av_packet_unref(file->pkt);
        int err = av_read_frame(file->format, file->pkt);
        if (err < 0) {
            // TODO: Error message.
            return err;
        }

        if (file->pkt->stream_index != file->stream) {
            continue;
        }

        err = avcodec_send_packet(file->codec, file->pkt);
        // TODO: Error message.

        return err;
    }
}

// Decode single frame and store decoded data in file->buf.
static int ffmpeg_decode_frame(struct ffmpeg_file *file)
{
    AVFrame *frame = file->frame;
    AVCodecContext *codec = file->codec;

    for (;;) {
        av_frame_unref(file->frame);
        int err = avcodec_receive_frame(codec, frame);
        if (err == AVERROR(EAGAIN)) {
            err = ffmpeg_send_packet(file);
            if (err == AVERROR_EOF) {
                return 0; // TODO: Return number of decoded.
            }
            if (err < 0) {
                return err;
            }
            continue;
        }
        if (err == AVERROR_EOF) {
            return 0; // TODO: Return number of decoded bytes.
        }
        if (err < 0) {
            // TODO: Error string.
            return err;
        }

        break;
    }

    int delay_nsamples = swr_get_delay(file->swr, codec->sample_rate);
    int dst_nsamples = av_rescale_rnd(delay_nsamples + frame->nb_samples,
        file->sample_rate, file->codec->sample_rate, AV_ROUND_UP);

    if (file->buf_nsamples < dst_nsamples) {
        if (file->buf) {
            av_freep(&file->buf[0]);
        }
        av_freep(&file->buf);
        int e = av_samples_alloc_array_and_samples(&file->buf, NULL,
                file->channels, file->sample_rate, file->sample_fmt, 0);
        if (e < 0) {
            return e;
        }
        file->buf_nsamples = dst_nsamples;
    }

    int ns = swr_convert(file->swr, file->buf, dst_nsamples,
        (const uint8_t**) frame->data, frame->nb_samples);
    int nb = av_samples_get_buffer_size(NULL, file->channels,
            ns, file->sample_fmt, 1);
    if (nb < 0) {
        return nb;
    }
    file->buf_len = nb;
    file->buf_offset = 0;
    if (file->frame->pts > 0) {
        file->time = file->frame->pts;
    }

    return nb;
}

// Return string representation of provided FFmpeg error code.
// It is client's responsibility to free memory allocated by string.
char *ffmpeg_strerror(int err)
{
    char *buf = malloc(AV_ERROR_MAX_STRING_SIZE);
    int e = av_strerror(err, buf, AV_ERROR_MAX_STRING_SIZE);
    if (e < 0) {
        strncpy(buf, "not ffmpeg error", AV_ERROR_MAX_STRING_SIZE);
    }

    return buf;
}

void ffmpeg_init()
{
    av_log_set_level(AV_LOG_ERROR);
}

void ffmpeg_metadata_free(struct ffmpeg_metadata *md)
{
    if (md->artist) {
        free(md->artist);
    }
    if (md->album) {
        free(md->album);
    }
    if (md->title) {
        free(md->title);
    }
    if (md->number) {
        free(md->number);
    }
    free(md);
}

// Allocate and initialize ffmpeg structure.
// ffmpeg_free() must be called to free allocated memory.
struct ffmpeg_file *ffmpeg_alloc()
{
    return zmalloc(sizeof(struct ffmpeg_file));
}

// Free structure allocated by ffmpeg_alloc().
void ffmpeg_free(struct ffmpeg_file *file)
{
    free(file);
}

// Open audio file, associate it with ffmpeg structure and prepare
// for reading its data.
int ffmpeg_open(struct ffmpeg_file *file, const char *filename)
{
    file->format = avformat_alloc_context();
    int err = avformat_open_input(&file->format, filename, NULL, NULL);
    if (err != 0) {
        ffmpeg_close(file);

        return err;
    }
    err = avformat_find_stream_info(file->format, NULL);
    if (err < 0) {
        ffmpeg_close(file);

        return err;
    }

    file->stream = -1;
    for (int i = 0; i < file->format->nb_streams; i++) {
        if (file->format->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_AUDIO) {
            file->stream = i;
            break;
        }
    }
    if (file->stream == -1) {
        ffmpeg_close(file);

        return AVERROR_STREAM_NOT_FOUND;
    }

    file->channels = 2;
    file->sample_rate = 44100;
    file->sample_fmt = AV_SAMPLE_FMT_S16;

    return 0;
}

// Close ffmpeg structure openned by ffmpeg_open().
void ffmpeg_close(struct ffmpeg_file *file)
{
    if (file->buf) {
        av_freep(&file->buf[0]);
        av_freep(&file->buf);
    }
    if (file->swr) {
        swr_free(&file->swr);
    }
    if (file->pkt) {
        av_packet_unref(file->pkt);
        av_packet_free(&file->pkt);
    }
    if (file->frame) {
        av_frame_unref(file->frame);
        av_frame_free(&file->frame);
    }
    if (file->codec) {
        avcodec_close(file->codec);
        avcodec_free_context(&file->codec);
    }
    if (file->format) {
        avformat_close_input(&file->format);
        avformat_free_context(file->format);
    }
}

struct ffmpeg_metadata *ffmpeg_metadata(struct ffmpeg_file *file)
{
    AVStream *s = file->format->streams[file->stream];
    struct ffmpeg_metadata *md = zmalloc(sizeof(struct ffmpeg_metadata));
    md->duration = (int) (av_q2d(s->time_base) * s->duration);

    AVDictionary *m = file->format->metadata;
    AVDictionaryEntry *tag = NULL;
    while ((tag = av_dict_get(m, "", tag, AV_DICT_IGNORE_SUFFIX))) {
        if (strcasecmp(tag->key, "artist") == 0) {
            md->artist = strdup(tag->value);
        } else if (strcasecmp(tag->key, "album") == 0) {
            md->album = strdup(tag->value);
        } else if (strcasecmp(tag->key, "date") == 0) {
            md->year = atoi(tag->value);
        } else if (strcasecmp(tag->key, "title") == 0) {
            md->title = strdup(tag->value);
        } else if (strcasecmp(tag->key, "track") == 0) {
            md->number = strdup(tag->value);
        }
    }

    return md;
}

int ffmpeg_open_codec(struct ffmpeg_file *file)
{
    const AVStream *s = file->format->streams[file->stream];
    const AVCodec *decoder = avcodec_find_decoder(s->codecpar->codec_id);
    AVCodecContext *codec = avcodec_alloc_context3(decoder);

    int err = avcodec_parameters_to_context(codec,
        file->format->streams[file->stream]->codecpar);
    if (err < 0) {
        // TODO: Set error message.
        return -1;
    }

    if (avcodec_open2(codec, decoder, NULL) < 0) {
        // TODO: Set error message.
        return -1;
    }
    file->codec = codec;

    file->pkt = av_packet_alloc();
    av_init_packet(file->pkt);
    av_packet_unref(file->pkt);

/*     file->time = 0; */
    file->frame = av_frame_alloc();
    if (!file->frame) {
        // TODO: Set error message.
        return -1;
    }

    file->swr = swr_alloc();
    if (!file->swr) {
        return -1;
    }
    av_opt_set_int(file->swr, "in_channel_count",
            file->codec->channels, 0);
    av_opt_set_int(file->swr, "out_channel_count",
            file->channels, 0);
    av_opt_set_int(file->swr, "in_channel_layout",
            file->codec->channel_layout, 0);
    av_opt_set_int(file->swr, "out_channel_layout",
            AV_CH_LAYOUT_STEREO, 0);
    av_opt_set_int(file->swr, "in_sample_rate",
            file->codec->sample_rate, 0);
    av_opt_set_int(file->swr, "out_sample_rate",
            file->sample_rate, 0);
    av_opt_set_sample_fmt(file->swr, "in_sample_fmt",
            file->codec->sample_fmt, 0);
    av_opt_set_sample_fmt(file->swr, "out_sample_fmt",
            file->sample_fmt,  0);
    swr_init(file->swr);
    if (!swr_is_initialized(file->swr)) {
        return -1;
    }

    return 0;
}

// Decode next len bytes. Returns number of bytes decoded, zero on
// stream end and negative number in case of error.
int ffmpeg_read(struct ffmpeg_file *file, char *buf, int len)
{
    int wrote = 0;
    int err;

    for (;;) {
        int left = len - wrote;
        int nready = MIN(left, file->buf_len - file->buf_offset);
        if (nready) {
            memcpy(buf + wrote, file->buf[0] + file->buf_offset, nready);
            file->buf_offset += nready;
            wrote += nready;
        }

        if (wrote == len) {
            break;
        }

        int n = ffmpeg_decode_frame(file);
        if (n <= 0) {
            return n;
        }
    }

    return wrote;
}

int ffmpeg_seek(struct ffmpeg_file *file, int pos)
{
    if (pos < 0) {
        return -1;
    }

    AVStream *s = file->format->streams[file->stream];
    int64_t delta_pts = av_rescale_q(pos, av_make_q(1, 1), s->time_base);
    int64_t pts = s->start_time + delta_pts;
    int e = av_seek_frame(file->format, file->stream, pts,
            AVSEEK_FLAG_ANY | AVSEEK_FLAG_BACKWARD);
    if (e < 0) {
        return e;
    }

    avcodec_flush_buffers(file->codec);
    file->time = pts;
    file->buf_len = 0;
    file->buf_offset = 0;

    return 0;
}

int ffmpeg_time(struct ffmpeg_file *file)
{
    AVStream *s = file->format->streams[file->stream];
    int ts = file->time * av_q2d(s->time_base);

    return ts;
}

int ffmpeg_channels(struct ffmpeg_file *file)
{
    return file->channels;
}

int ffmpeg_sample_rate(struct ffmpeg_file *file)
{
    return file->sample_rate;
}
