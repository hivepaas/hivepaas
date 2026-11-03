/**
 * A duration as the server reads one: amounts with their units, run together -
 * 1d, 12h, 1h30m, 90d - with days (d) and weeks (w) besides Go's own units. One
 * it cannot read fails the whole request, naming no field, so forms check it
 * first.
 */
const DURATION_PATTERN = /^(?:\d+(?:\.\d+)?(?:ns|us|µs|ms|s|m|h|d|w))+$/;

export const DURATION_HINT = "Use a duration such as 1d, 12h or 1h30m";

export function isDuration(value: string): boolean {
    return DURATION_PATTERN.test(value.trim());
}
