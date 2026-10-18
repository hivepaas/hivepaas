import { format } from "date-fns";
import type { AppManagedSettings } from "~/projects/domain";

/** A moment as the managed screens show it: to the minute, on this browser's clock. */
export function formatManagedTime(date: Date): string {
    return format(date, "yyyy-MM-dd HH:mm");
}

/** A restored database is restoring until its engine has seen it recover and back up. */
export function isRestoring(settings: AppManagedSettings): boolean {
    return settings.restore !== undefined && settings.restore.readyAt === undefined;
}

const DAY_MS = 24 * 60 * 60 * 1000;

/**
 * About how much more the archive keeps for a point kept until then, at one base
 * backup a night: the base backups retention would have deleted by then, each of
 * about the last one's size. The WAL written meanwhile is kept too, and is not
 * counted: null when there is no base backup to go by.
 */
export function estimateKeptBytes(
    keepUntil: Date,
    now: Date,
    retainFull: number,
    lastBackupBytes: number | undefined,
): { backups: number; bytes: number } | null {
    if (lastBackupBytes === undefined || lastBackupBytes <= 0) {
        return null;
    }
    const nights = Math.ceil((keepUntil.getTime() - now.getTime()) / DAY_MS);
    const backups = Math.max(0, nights - retainFull);

    return { backups, bytes: backups * lastBackupBytes };
}

/** The most a point may be kept for, as the server allows. */
export const MAX_POINT_KEEP_DAYS = 90;

/** A restore point's name, as the engine takes it: Postgres's 63 bytes, and not one of HivePaaS's own. */
export const POINT_NAME_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]{0,62}$/;
export const RESERVED_POINT_PREFIX = "hp-";
