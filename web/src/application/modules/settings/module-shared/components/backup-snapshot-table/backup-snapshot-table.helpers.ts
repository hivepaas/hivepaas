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
