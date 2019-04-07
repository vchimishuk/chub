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

static char *ffmpeg_last_err = NULL;

static void *ffmpeg_alloc(size_t n)
{
    void *p = malloc(n);
    memset(p, 0, n);

    return p;
}

static void ffmpeg_reset_pkt(struct ffmpeg_file *file)
{
    av_packet_unref(file->pkt);
    file->pkt_decoded = file->pkt->size;
}

static int ffmpeg_decode(struct ffmpeg_file *file)
{
    if (file->pkt_decoded >= file->pkt->size) {
        int e = av_read_frame(file->format, file->pkt);
        if (e < 0) {
            return e;
        }
        file->pkt_decoded = 0;
    }

    int got_frame = 0;
    int e = avcodec_decode_audio4(file->codec, file->frame, &got_frame,
            file->pkt);
    if (e < 0) {
        return e;
    }
    if (!got_frame) {
        return 0;
    }
    file->pkt_decoded += e;

    AVFrame *frame = file->frame;
    int delay_nsamples = swr_get_delay(file->swr, file->codec->sample_rate);
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
    file->time = file->frame->pts;

    return nb;
}

char *ffmpeg_last_error()
{
    // TODO: Implement.
    // TODO: Return (or try to get from ffmpeg) string representation of errors.
    return ffmpeg_last_err;
}

void ffmpeg_init()
{
    // TODO: Register only used audio formats.
    av_log_set_level(AV_LOG_ERROR);
    av_register_all();
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

struct ffmpeg_file *ffmpeg_open(const char *filename)
{
    struct ffmpeg_file *f = ffmpeg_alloc(sizeof(struct ffmpeg_file));

    f->format = avformat_alloc_context();
    int err = avformat_open_input(&f->format, filename, NULL, NULL);
    if (err != 0) {
        // TODO: Set last error string.
        ffmpeg_close(f);

        return NULL;
    }
    err = avformat_find_stream_info(f->format, NULL);
    if (err < 0) {
        // TODO: Set last error string.
        ffmpeg_close(f);

        return NULL;
    }

    f->stream = -1;
    for (int i = 0; i < f->format->nb_streams; i++) {
        if (f->format->streams[i]->codec->codec_type == AVMEDIA_TYPE_AUDIO) {
            f->stream = i;
            break;
        }
    }
    if (f->stream == -1) {
        // TODO: Set last error string.
        ffmpeg_close(f);

        return NULL;
    }

    f->channels = 2;
    f->sample_rate = 44100;
    f->sample_fmt = AV_SAMPLE_FMT_S16;

    return f;
}

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
        av_free_packet(file->pkt);
    }
    if (file->frame) {
        av_frame_free(&file->frame);
    }
    if (file->codec) {
        avcodec_close(file->codec);
    }
    if (file->format) {
        avformat_close_input(&file->format);
    }
}

struct ffmpeg_metadata *ffmpeg_metadata(struct ffmpeg_file *file)
{
    AVStream *s = file->format->streams[file->stream];
    struct ffmpeg_metadata *md = ffmpeg_alloc(sizeof(struct ffmpeg_metadata));
    md->duration = (int) (av_q2d(s->time_base) * s->duration);

    AVDictionary *m = file->format->metadata;
    AVDictionaryEntry *tag = NULL;
    while ((tag = av_dict_get(m, "", tag, AV_DICT_IGNORE_SUFFIX))) {
        if (strcmp(tag->key, "artist") == 0) {
            md->artist = strdup(tag->value);
        } else if (strcmp(tag->key, "album") == 0) {
            md->album = strdup(tag->value);
        } else if (strcmp(tag->key, "title") == 0) {
            md->title = strdup(tag->value);
        } else if (strcmp(tag->key, "track") == 0) {
            md->number = strdup(tag->value);
        }
    }

    return md;
}

int ffmpeg_open_codec(struct ffmpeg_file *file)
{
    AVStream *s = file->format->streams[file->stream];
    AVCodec *decoder = avcodec_find_decoder(s->codec->codec_id);

    if (avcodec_open2(s->codec, decoder, NULL) < 0) {
        return -1;
    }
    file->codec = s->codec;
    file->pkt = av_malloc(sizeof(AVPacket));
    av_init_packet(file->pkt);
    av_packet_unref(file->pkt);

    file->time = 0;
    file->frame = av_frame_alloc();
    if (!file->frame) {
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

int ffmpeg_read(struct ffmpeg_file *file, char *buf, int len)
{
    int wrote = 0;

    while (len - wrote > 0) {
        if (file->buf == NULL || file->buf_len - file->buf_offset == 0) {
            for (int i = 0; ; i++) {
                int n = ffmpeg_decode(file);
                if (n < 0) {
                    ffmpeg_reset_pkt(file);
                    if (i >= 3) {
                        return n;
                    }
                } else {
                    break;
                }
            }
        }

        if (file->buf_len - file->buf_offset > 0) {
            int n = MIN(len - wrote, file->buf_len - file->buf_offset);
            memcpy(buf + wrote, file->buf[0] + file->buf_offset, n);
            file->buf_offset += n;
            wrote += n;
        }
    }

    return wrote;
}

int ffmpeg_seek(struct ffmpeg_file *file, int pos, int rel)
{
    if (!rel && pos < 0) {
        return -1;
    }

    AVStream *s = file->format->streams[file->stream];
    int64_t pos_pts = pos / av_q2d(s->time_base);
    int pts;
    if (rel) {
        pts = file->time + pos_pts;
    } else {
        pts = s->start_time + pos_pts;
    }
    int e = av_seek_frame(file->format, file->stream, pts,
            AVSEEK_FLAG_ANY | AVSEEK_FLAG_BACKWARD);
    if (e < 0) {
        return e;
    }

    avcodec_flush_buffers(file->codec);
    file->time = pts;
    file->buf_len = 0;
    file->buf_offset = 0;
    ffmpeg_reset_pkt(file);

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
