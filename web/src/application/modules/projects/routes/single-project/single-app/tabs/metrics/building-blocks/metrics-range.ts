import type { FunctionMetricsRange } from "~/projects/api/services";

export const METRICS_RANGES: { value: FunctionMetricsRange; label: string }[] = [
    { value: "1h", label: "1 hour" },
    { value: "6h", label: "6 hours" },
    { value: "24h", label: "24 hours" },
    { value: "7d", label: "7 days" },
];

/** A day: a function's busy hours, its scheduled calls, what failed overnight. */
export const DEFAULT_METRICS_RANGE: FunctionMetricsRange = "24h";

const STORAGE_KEY = "hivepaas.function-metrics.range";

/** The range last chosen in this browser; the default when there is none, or no storage. */
export function storedMetricsRange(): FunctionMetricsRange {
    try {
        const stored = window.localStorage.getItem(STORAGE_KEY);
        return METRICS_RANGES.find(range => range.value === stored)?.value ?? DEFAULT_METRICS_RANGE;
    } catch {
        return DEFAULT_METRICS_RANGE;
    }
}

export function storeMetricsRange(range: FunctionMetricsRange): void {
    try {
        window.localStorage.setItem(STORAGE_KEY, range);
    } catch {
        // A browser without storage starts at the default next time.
    }
}

/** A point's time on the axis: hours and minutes, with the day for a week. */
export function formatMetricsTime(time: string, range: FunctionMetricsRange): string {
    const at = new Date(time);
    const hm = at.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    return range === "7d" ? `${at.toLocaleDateString([], { month: "short", day: "numeric" })} ${hm}` : hm;
}
