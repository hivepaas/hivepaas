/**
 * LOG_FILTER_FIELD_WIDTH is the width every log filter control shares.
 *
 * Since, Duration, Levels and Streams sit in one row and read as one group, so
 * they are one size. A label too long for it truncates rather than pushing the
 * row out of alignment.
 */
export const LOG_FILTER_FIELD_WIDTH = "w-[88px] sm:w-[120px]";
