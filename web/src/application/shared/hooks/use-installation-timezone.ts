import { SessionQueries } from "@application/shared/data/queries";

/**
 * The installation's timezone, a zone name such as America/New_York: what a schedule's hours are read in - a cron
 * expression's, the time of day a system job runs at. From the profile the dashboard loaded at sign-in; UTC until
 * it is there.
 */
export function useInstallationTimezone(): string {
    const { data } = SessionQueries.useGetProfile({ staleTime: Infinity });
    return data?.data.timezone ?? "UTC";
}

/**
 * The timezone times can be shown in here: the one given, or UTC for a zone this browser does not know.
 */
export function shownTimezone(timeZone: string): string {
    try {
        new Intl.DateTimeFormat("en-US", { timeZone });
        return timeZone;
    } catch {
        return "UTC";
    }
}

/**
 * A time as yyyy-MM-dd HH:mm:ss on the clocks of a timezone; in UTC for a zone this browser does not know.
 */
export function formatInTimezone(date: Date, timeZone: string): string {
    const options: Intl.DateTimeFormatOptions = {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hourCycle: "h23",
    };
    try {
        // Swedish writes a date and time as ISO 8601 does.
        return new Intl.DateTimeFormat("sv-SE", { ...options, timeZone }).format(date);
    } catch {
        return new Intl.DateTimeFormat("sv-SE", { ...options, timeZone: "UTC" }).format(date);
    }
}
