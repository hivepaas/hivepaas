/** What the live stream offers. A container rarely holds more than a week. */
export const DEFAULT_DURATION_OPTIONS = ["5m", "15m", "30m", "1h", "2h", "4h", "8h", "1d", "2d", "4d", "7d"] as const;

/**
 * What stored logs offer: the same shortcuts plus the ones a long retention
 * makes reachable. The text field takes anything either way, so the list is a
 * convenience, not a limit.
 */
export const HISTORY_DURATION_OPTIONS = [...DEFAULT_DURATION_OPTIONS, "14d", "30d", "90d"] as const;
