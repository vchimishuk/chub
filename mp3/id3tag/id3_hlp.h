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

#ifndef ID3_HLP_H
#define ID3_HLP_H

#include <id3tag.h>

/*
 * Returns tag's frame by given frame number.
 */
struct id3_frame *id3_hlp_get_tag_frame(struct id3_tag *tag,
                                        unsigned int frame_num);

/*
 * Returns frame's ID string.
 */
char *id3_hlp_get_frame_id(struct id3_frame *frame);

/*
 * Returns type for given frame.
 */
enum id3_field_type id3_hlp_get_frame_type(struct id3_frame *frame);

/*
 * Returns frames string value.
 */
char *id3_hlp_get_frame_string(struct id3_frame *frame);

#endif // ID3_HLP_H
