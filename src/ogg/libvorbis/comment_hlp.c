#include <stdlib.h>
#include "comment_hlp.h"

char *comment_hlp_get_user_comment(const vorbis_comment *comment, int i)
{
        if (i < 0 || i >= comment->comments) {
                return NULL;
        }

        return comment->user_comments[i];
}
