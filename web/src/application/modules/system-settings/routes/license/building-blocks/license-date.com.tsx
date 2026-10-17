import { formatLicenseDate, formatLicenseDateTime } from "@application/shared/utils";

/**
 * A license date: the day, with the hour it falls at in its title, as a key
 * runs out at an hour rather than on a day.
 */
export function LicenseDate({ date }: { date: Date }) {
    return (
        <time
            dateTime={date.toISOString()}
            title={formatLicenseDateTime(date)}
        >
            {formatLicenseDate(date)}
        </time>
    );
}
