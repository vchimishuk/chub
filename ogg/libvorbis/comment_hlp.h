#ifndef CHUB_COMMENT_HLP_H
#define CHUB_COMMENT_HLP_H

#include <vorbis/codec.h>
#include <vorbis/vorbisfile.h>

// TODO: Rename all OGG C functions: add "ogg_" prefix.

/*
 * Returns user_comment string by its index.
 */
char *comment_hlp_get_user_comment(const vorbis_comment *comment, int i);

#endif // CHUB_COMMENT_HLP_H
