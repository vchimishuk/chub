#include <stdlib.h>
#include "id3_hlp.h"


struct id3_frame *id3_hlp_get_tag_frame(struct id3_tag *tag,
                                        unsigned int frame_num)
{
        if (frame_num >= tag->nframes) {
                return NULL;
        }

        return tag->frames[frame_num];
}

char *id3_hlp_get_frame_id(struct id3_frame *frame)
{
        return frame->id;
}

enum id3_field_type id3_hlp_get_frame_type(struct id3_frame *frame)
{
        return id3_field_type(&frame->fields[1]);
}

char *id3_hlp_get_frame_string(struct id3_frame *frame)
{
        char *str;

        if (id3_field_getnstrings(&frame->fields[1]) != 0) {
                str = (char *) id3_field_getstrings(&frame->fields[1], 0);
                if (str != NULL) {
                        str = id3_ucs4_utf8duplicate((id3_ucs4_t const *) str);
                }
        }

        return str;
}
