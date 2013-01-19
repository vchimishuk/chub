#include <id3tag.h>

/*
 * Returns tag's frame by given frame number.
 */
struct id3_frame *id3_hlp_get_tag_frame(struct id3_tag *tag, unsigned int frame_num);

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
