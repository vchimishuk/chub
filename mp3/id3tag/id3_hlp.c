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
