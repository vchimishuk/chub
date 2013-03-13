#include <vorbis/codec.h>
#include <vorbis/vorbisfile.h>

/*
 * Returns user_comment string by its index.
 */
char *comment_hlp_get_user_comment(const vorbis_comment *comment, int i);
