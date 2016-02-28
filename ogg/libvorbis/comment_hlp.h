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

#ifndef COMMENT_HLP_H
#define COMMENT_HLP_H

#include <vorbis/codec.h>
#include <vorbis/vorbisfile.h>

// TODO: Rename all OGG C functions: add "ogg_" prefix.

/*
 * Returns user_comment string by its index.
 */
char *comment_hlp_get_user_comment(const vorbis_comment *comment, int i);

#endif // COMMENT_HLP_H
