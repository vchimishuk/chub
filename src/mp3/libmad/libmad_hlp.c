#include "libmad_hlp.h"


/*
 * Convert a fixed point sample to a short.
 */
static inline int16_t gomad_fixed_to_short(mad_fixed_t fixed)
{
	if (fixed >= MAD_F_ONE) {
		return SHRT_MAX;
	} else if (fixed <= -MAD_F_ONE) {
		return -SHRT_MAX;
	}
	
	return fixed >> (MAD_F_FRACBITS - 15);
}


/*
 * Reinitialize decoder, so file will be decoded from the beginning.
 */
static void gomad_rewind(struct gomad_decoder *decoder)
{
	rewind(decoder->file);

	decoder->current_position = 0;
	decoder->current_sample = 0;
	
	mad_stream_init(&decoder->stream);
	mad_frame_init(&decoder->frame);
	mad_header_init(&decoder->header);
	mad_synth_init(&decoder->synth);
	mad_timer_reset(&decoder->timer);
}


/*
 * Fill buffer with new data from input file.
 * Returns size of actual data in the buffer.
 */
static size_t gomad_fill_buffer(struct gomad_decoder *decoder)
{
	size_t offset;
	size_t free_size;
	size_t read_len;
		
	if (decoder->stream.next_frame == NULL) {
		offset = 0;
	} else {
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
static int gomad_read_frame(struct gomad_decoder *decoder)
{
	size_t data_size;

	for (; ;) {
		if (decoder->stream.buffer == NULL || decoder->stream.error == MAD_ERROR_BUFLEN) {
			data_size = gomad_fill_buffer(decoder);
			if (data_size <= 0) {
				return -1;
			}

			mad_stream_buffer(&decoder->stream, decoder->buf, data_size);
			decoder->stream.error = MAD_ERROR_NONE;
		}

		if (mad_header_decode(&decoder->header, &decoder->stream) == 0) {
			mad_timer_add(&decoder->timer, decoder->frame.header.duration);
			decoder->current_position = decoder->timer.seconds;
			decoder->frame.header = decoder->header;
			return 0;
		}
		
		if (!MAD_RECOVERABLE(decoder->stream.error) && (decoder->stream.error != MAD_ERROR_BUFLEN)) {
			return -1;
		}
	}
}


/*
 * Calculate and returns length of the file in seconds.
 */
static void gomad_fill_info(struct gomad_decoder *decoder)
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
		if (gomad_read_frame(decoder) != 0) {
			decoder->length = decoder->timer.seconds;
			offset = ftello(decoder->file);
			gomad_rewind(decoder);

			return;
		}

		if (first_frame) {
			if (mad_frame_decode(&decoder->frame, &decoder->stream) != -1) {
				decoder->sample_rate = decoder->frame.header.samplerate;
				decoder->channels = MAD_NCHANNELS(&decoder->frame.header);
				first_frame = 0;
			}
		}
	} while (decoder->timer.seconds < 100);

	/* Extrapolate the time. Not really accurate, but good enough */
	offset = ftello(decoder->file);
	fseek(decoder->file, 0, SEEK_END);
	total = ftello(decoder->file);

	decoder->length = ((double) total / offset) * decoder->timer.seconds;

	gomad_rewind(decoder);
	
	return;
}


int gomad_open(struct gomad_decoder *decoder, const char *filename)
{
	decoder->file = fopen(filename, "r");
	if (decoder->file == NULL) {
		return -1;
	}

	gomad_rewind(decoder);
	gomad_fill_info(decoder);
	
	return 0;
}


size_t gomad_read(struct gomad_decoder *decoder, char *buf, size_t len)
{
	int16_t *words_buf = (int16_t *)buf;
	size_t words_len = len / 2;

	size_t written = 0;
	int i;

	do {
		if (decoder->current_sample == 0) {
			if (gomad_read_frame(decoder) == -1) {
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
				words_buf[written++] = gomad_fixed_to_short(decoder->synth.pcm.samples[i][decoder->current_sample]);

			}

			decoder->current_sample++;
		}

		if (decoder->current_sample == decoder->synth.pcm.length) {
			decoder->current_sample = 0;
		}
	} while (written < words_len);

	return written * 2;
}


void gomad_close(struct gomad_decoder *decoder)
{
	mad_frame_finish(&decoder->frame);
	mad_stream_finish(&decoder->stream);
	mad_synth_finish(&decoder->synth);

	fclose(decoder->file);
}
