import { format } from "date-fns";
import type { ClusterVolume } from "~/cluster/domain";
import type { AppManagedSettings } from "~/projects/domain";

/** A moment as the managed screens show it: to the minute, on this browser's clock. */
export function formatManagedTime(date: Date): string {
    return format(date, "yyyy-MM-dd HH:mm");
}

/** Where a volume's data is, as a person reads it: a node, the nodes with a label, or every node - shared storage. */
export function volumePlaceOf(volume: Pick<ClusterVolume, "nodeId" | "nodeLabel">): string {
    if (volume.nodeId) {
        return `node ${volume.nodeId}`;
    }
    return volume.nodeLabel ? `nodes labeled ${volume.nodeLabel}` : "every node";
}

/** A tmpfs volume keeps its data in memory, gone at each restart: never a database's data. */
export function isVolumeInMemory(volume: Pick<ClusterVolume, "tmpfsOptions">): boolean {
    return Boolean(volume.tmpfsOptions);
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

/** An image reference without its tag or digest: "postgres" of "postgres:18.6". */
export function imageRepositoryOf(image: string): string {
    const withoutDigest = image.split("@")[0] ?? image;
    const lastSlash = withoutDigest.lastIndexOf("/");
    const lastColon = withoutDigest.lastIndexOf(":");

    return lastColon > lastSlash ? withoutDigest.slice(0, lastColon) : withoutDigest;
}
