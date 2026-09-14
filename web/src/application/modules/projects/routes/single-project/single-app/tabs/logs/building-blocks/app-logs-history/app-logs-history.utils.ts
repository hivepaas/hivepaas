const SECOND_MS = 1000;
const MINUTE_MS = 60 * SECOND_MS;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;
const WEEK_MS = 7 * DAY_MS;

const DURATION_UNIT_MS: Record<string, number> = {
    w: WEEK_MS,
    d: DAY_MS,
    h: HOUR_MS,
    m: MINUTE_MS,
    s: SECOND_MS,
};

/** The units the server's duration parser accepts, in the order it writes them. */
const DURATION_PATTERN = /^(?:\d+(?:w|d|h|m|s))+$/;
const DURATION_PART_PATTERN = /(\d+)(w|d|h|m|s)/g;

/**
 * The window used when neither Since nor Duration is set.
 *
 * It mirrors the server's own default (`defaultLogHistoryWindow`), but stored
 * logs still send it explicitly: see resolveLogHistoryWindow.
 */
export const DEFAULT_LOG_HISTORY_WINDOW_MS = HOUR_MS;

/**
 * parseDurationMs reads a duration such as `30m`, `7d` or `1d12h`.
 *
 * Anything the server would reject reads as undefined here, so a half-typed
 * duration leaves the window alone instead of narrowing it to nothing.
 */
export function parseDurationMs(value: string | undefined): number | undefined {
    const normalizedValue = value?.trim();

    if (!normalizedValue || !DURATION_PATTERN.test(normalizedValue)) {
        return undefined;
    }

    let total = 0;

    for (const [, amount, unit] of normalizedValue.matchAll(DURATION_PART_PATTERN)) {
        const unitMs = unit === undefined ? undefined : DURATION_UNIT_MS[unit];

        if (unitMs === undefined) {
            return undefined;
        }

        total += Number(amount) * unitMs;
    }

    return total > 0 ? total : undefined;
}

/**
 * resolveLogHistoryWindow turns the toolbar's Since and Duration into the
 * `start` and `end` the stored-logs endpoint takes.
 *
 * The three cases are the live stream's, so that one toolbar means one thing on
 * both tabs (see GetAppLogs in the server's logs_get.go):
 *
 *   duration only  -> the last N, up to now
 *   since only     -> from then until now
 *   since+duration -> a fixed window starting at Since, never past now
 *
 * `start` is always returned, and always sent. Leaving it out would let the
 * server derive it from `end` - which paging walks backwards - and the window
 * would slide open one page at a time instead of ending.
 *
 * `endsNow` marks the two thirds of those cases whose end is simply "now". Such
 * a window goes stale the moment it is resolved, and Refresh is what re-opens
 * it; a window a person closed by hand does not move.
 */
export function resolveLogHistoryWindow(
    since: Date | undefined,
    duration: string | undefined,
    now: Date,
): LogHistoryWindow {
    const durationMs = parseDurationMs(duration);

    if (!since) {
        return {
            start: new Date(now.getTime() - (durationMs ?? DEFAULT_LOG_HISTORY_WINDOW_MS)),
            end: now,
            endsNow: true,
        };
    }
    if (durationMs === undefined) {
        return { start: since, end: now, endsNow: true };
    }

    // A window running past now is clamped back to now, unless Since is itself
    // that late: the server rejects a start that is not before the end, and a
    // duration is always positive, so the unclamped end is the safe fallback.
    const requestedEnd = since.getTime() + durationMs;
    const clampedEnd = Math.min(requestedEnd, now.getTime());
    const isClamped = clampedEnd < requestedEnd && clampedEnd > since.getTime();

    return {
        start: since,
        end: new Date(isClamped ? clampedEnd : requestedEnd),
        // Clamped means the window ran past now and now is what ended it.
        endsNow: isClamped,
    };
}

export interface LogHistoryWindow {
    start: Date;
    end: Date;
    /** The end is "now" as of when this was resolved, not a bound a person set. */
    endsNow: boolean;
}

/**
 * resolveDurationOptions narrows the duration shortcuts to what the store
 * actually holds.
 *
 * Offering `90d` against a seven-day retention is an offer of nothing. The
 * retention itself is appended when no shortcut reaches that far, so the whole
 * depth is always one click away, and an unknown retention (an external
 * backend) leaves the list as it was.
 */
export function resolveDurationOptions(options: readonly string[], retention: string | undefined): readonly string[] {
    const retentionMs = parseDurationMs(retention);

    if (retentionMs === undefined) {
        return options;
    }

    const normalizedRetention = retention?.trim() ?? "";
    const withinRetention = options.filter(option => (parseDurationMs(option) ?? 0) <= retentionMs);

    return withinRetention.includes(normalizedRetention) ? withinRetention : [...withinRetention, normalizedRetention];
}

/**
 * earliestStoredLogTime is the oldest instant Since can usefully point at.
 *
 * Undefined means the depth is unknown, and the picker then bounds nothing.
 */
export function earliestStoredLogTime(retention: string | undefined, now: Date): Date | undefined {
    const retentionMs = parseDurationMs(retention);

    return retentionMs === undefined ? undefined : new Date(now.getTime() - retentionMs);
}

/**
 * describeLogHistoryView says what is on screen and, more to the point, what is
 * not.
 *
 * The window and the line count bound the view independently, and neither is
 * visible in the result: 500 lines out of a 90-day window look exactly like 500
 * lines out of a 15-minute one. Without this a person reasonably reads an
 * unfinished view as the whole of the range they asked for.
 */
export function describeLogHistoryView({
    window: timeWindow,
    lineCount,
    oldestLoaded,
    hasMore,
    isLoading,
}: LogHistoryViewDescription): string {
    const range = `${formatWindowInstant(timeWindow.start)} → ${timeWindow.endsNow ? "now" : formatWindowInstant(timeWindow.end)}`;

    if (isLoading) {
        return `${range} · loading…`;
    }
    if (lineCount === 0) {
        return `${range} · no lines`;
    }

    const loaded = `${lineCount.toLocaleString()} ${lineCount === 1 ? "line" : "lines"} loaded`;
    const oldest = oldestLoaded ? ` · oldest ${formatWindowInstant(oldestLoaded)}` : "";

    return `${range} · ${loaded}${oldest} · ${hasMore ? "more to load" : "start of range reached"}`;
}

/** Day and minute is as much as a status line can carry without becoming noise. */
function formatWindowInstant(value: Date): string {
    const pad = (part: number) => part.toString().padStart(2, "0");

    return `${pad(value.getDate())}/${pad(value.getMonth() + 1)} ${pad(value.getHours())}:${pad(value.getMinutes())}`;
}

export interface LogHistoryViewDescription {
    window: LogHistoryWindow;
    lineCount: number;
    oldestLoaded: Date | undefined;
    hasMore: boolean;
    isLoading: boolean;
}
