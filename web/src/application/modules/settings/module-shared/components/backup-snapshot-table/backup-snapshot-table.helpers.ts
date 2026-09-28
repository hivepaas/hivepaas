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

/** A source kind, as a person reads it. */
export function sourceLabel(source: string): string {
    switch (source) {
        case "command":
            return "Command";
        case "volume":
            return "Volume";
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
const KNOWN_TAG_CLASS_NAMES: Record<string, string> = {
    "hivepaas.app": "bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border-cyan-500/30",
    "hivepaas.job": "bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-500/30",
    "hivepaas.run": "bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/30",
    "hivepaas.source": "bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/30",
};

/** Every other tag - a job's own, or one set outside HivePaaS - shares one color. */
const OTHER_TAG_CLASS_NAME = "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/30";

/** A tag's badge colors, by its key. */
export function tagClassName(tag: string): string {
    const [key = ""] = tag.split(":");
    return KNOWN_TAG_CLASS_NAMES[key] ?? OTHER_TAG_CLASS_NAME;
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
