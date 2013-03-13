#include "ogg_hlp.h"


size_t ogg_hlp_read(OggVorbis_File *vf, char *buf, size_t size,
                    int bigendianp, int word, int sgned)
{
        int read = 0; /* Already read bytes. */
        
        while (read < size) {
                long rr = ov_read(vf, buf + read, size - read,
                                  bigendianp, word, sgned, NULL);
                if (rr <= 0) {
                        break;
                }

                read += rr;
        }

        return read;
}
