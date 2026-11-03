import type { BadgeTone } from "@components/ui/badge";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { MODULE_IDS } from "@application/shared/constants";

/** HivePaaS's own tags: where a snapshot came from, shown in their own columns. */
export const HIVEPAAS_TAG_PREFIX = "hivepaas.";

/** A snapshot's own tags, those its job was given. */
export function userTags(snapshot: BackupSnapshot): string[] {
    return snapshot.tags.filter(tag => !tag.startsWith(HIVEPAAS_TAG_PREFIX));
}

/** The module whose permission a view's actions ask for. */
export function snapshotScopeModuleId(scope: BackupSnapshotScope): string {
    return scope.type === "settings" ? MODULE_IDS.Settings : MODULE_IDS.Project;
}

/** The tag every system backup's snapshot carries. */
export const SYSTEM_BACKUP_SOURCE_TAG = "hivepaas.source:system-backup";

/**
 * Whether a snapshot can be restored from the dashboard. A system backup's cannot: its files are downloaded, and
 * HivePaaS's own database is restored by hand.
 */
export function canRestoreSnapshot(snapshot: BackupSnapshot): boolean {
    return snapshot.source !== "system-backup";
}

/** A source kind, as a person reads it. */
export function sourceLabel(source: string): string {
    switch (source) {
        case "command":
            return "Command";
        case "volume":
            return "Volume";
        case "system-backup":
            return "System";
        default:
            return "-";
    }
}

/** A tag filter is key:value. */
export function isTagFilter(value: string): boolean {
    const [key, ...rest] = value.split(":");
    return Boolean(key) && rest.length > 0;
}

/** The colors of the tags HivePaaS gives a snapshot, by key. */
const KNOWN_TAG_TONES: Record<string, BadgeTone> = {
    "hivepaas.app": "cyan",
    "hivepaas.job": "purple",
    "hivepaas.run": "amber",
    "hivepaas.source": "blue",
};

/** Every other tag - a job's own, or one set outside HivePaaS - shares one color. */
const OTHER_TAG_TONE: BadgeTone = "emerald";

/** A tag's badge color, by its key. */
export function tagTone(tag: string): BadgeTone {
    const [key = ""] = tag.split(":");
    return KNOWN_TAG_TONES[key] ?? OTHER_TAG_TONE;
}

export interface BackupSnapshotFilterValues {
    repo?: string;
    app?: string;
    tags: string[];
    fromDate?: string;
    toDate?: string;
}

/** How many filters are set, for the Filter button's count. */
export function countActiveFilters(filters: BackupSnapshotFilterValues): number {
    return [filters.repo, filters.app, filters.fromDate, filters.toDate].filter(Boolean).length + filters.tags.length;
}

/** A snapshot is read, to restore it, from a repository that is active. */
export function isRepoActive(snapshot: BackupSnapshot): boolean {
    return snapshot.repo.status === "active";
}
