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

#ifndef CHUB_FFMPEG_H
#define CHUB_FFMPEG_H

#include <libavformat/avformat.h>
#include <libswresample/swresample.h>

struct ffmpeg_metadata {
    char *artist;
    char *album;
    char *title;
    char *number;
    int duration;
};

struct ffmpeg_file {
    AVFormatContext *format;
    int stream;
    AVCodecContext *codec;
    AVPacket *pkt;
    AVFrame *frame;
    SwrContext *swr;
    int channels;
    int sample_rate;
    enum AVSampleFormat sample_fmt;
    /* Current decoding position in seconds. */
    int64_t time;
    uint8_t **buf;
    int buf_nsamples;
    int buf_len;
    int buf_offset;
    /* Number of bytes decoded from the current packet. */
    int pkt_decoded;
};

void ffmpeg_init();
char *ffmpeg_last_error();
struct ffmpeg_file *ffmpeg_open(const char *filename);
void ffmpeg_close(struct ffmpeg_file *file);
struct ffmpeg_metadata *ffmpeg_metadata(struct ffmpeg_file *file);
int ffmpeg_open_codec(struct ffmpeg_file *file);
int ffmpeg_read(struct ffmpeg_file *file, char *buf, int len);
int ffmpeg_seek(struct ffmpeg_file *file, int pos);
void ffmpeg_metadata_free(struct ffmpeg_metadata *metadata);
int ffmpeg_time(struct ffmpeg_file *file);
int ffmpeg_channels(struct ffmpeg_file *file);
int ffmpeg_sample_rate(struct ffmpeg_file *file);

#endif // CHUB_FFMPEG_H
