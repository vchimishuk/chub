#ifndef OGG_HLP_H
#defne OGG_HLP_H

#include <stdlib.h>
#include <vorbis/codec.h>
#include <vorbis/vorbisfile.h>

/*
 * Read size bytes from the stream into a buffer.
 * Returns actual number of read bytes.
 */
size_t ogg_hlp_read(OggVorbis_File *vf, char *buf, size_t size, int bigendianp, int word, int sgned);

#endif // OGG_HLP_H
