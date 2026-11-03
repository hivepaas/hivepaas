import type { BackupSnapshot, BackupSnapshotEntry, BackupSnapshotScope } from "~/settings/domain";

import { ROUTE } from "@application/shared/constants";

/** What a snapshot is restored as: a command's file, or a volume's directory. */
export interface RestoreKind {
    command: boolean;
    /** The file a command's snapshot holds. */
    fileName: string;
}

/**
 * A snapshot's kind: its source's, as its tag or its job says. One whose source is not known - taken outside
 * HivePaaS, or before the tag - goes by what it holds: one file is a command's. Undefined while its root is
 * still being listed.
 */
export function restoreKindOf(
    snapshot: BackupSnapshot,
    root: BackupSnapshotEntry[] | undefined,
): RestoreKind | undefined {
    if (snapshot.source === "volume") {
        return { command: false, fileName: "" };
    }
    const jobFileName = snapshot.job?.fileName ?? "";
    if (snapshot.source === "command" && jobFileName) {
        return { command: true, fileName: jobFileName };
    }
    if (!root) {
        return undefined;
    }
    const [first] = root;
    if (root.length === 1 && first && !first.dir) {
        return { command: true, fileName: first.name };
    }
    return { command: false, fileName: "" };
}

/** Whether the kind can be known without listing the snapshot. */
export function restoreKindNeedsListing(snapshot: BackupSnapshot): boolean {
    return snapshot.source !== "volume" && !(snapshot.source === "command" && snapshot.job?.fileName);
}

/** A path inside a snapshot, one directory deeper. */
export function joinSnapshotPath(parent: string, name: string): string {
    return parent ? `${parent}/${name}` : name;
}

/** The directory a path inside a snapshot is in; "" for the root's. */
export function parentSnapshotPath(path: string): string {
    const index = path.lastIndexOf("/");
    return index < 0 ? "" : path.slice(0, index);
}

/** Where a scope's snapshots are listed. */
export function snapshotListRoute(scope: BackupSnapshotScope): string {
    switch (scope.type) {
        case "settings":
            return ROUTE.settings.backupSnapshots.$route;
        case "project":
            return ROUTE.projects.single.providerConfiguration.backupSnapshots.$route(scope.projectId);
        case "app":
            return ROUTE.projects.single.apps.single.configuration.backupSnapshots.$route(
                scope.projectId,
                scope.env,
                scope.appId,
            );
    }
}

/** Where a scope's snapshot is restored. */
export function snapshotRestoreRoute(scope: BackupSnapshotScope, snapshotId: string): string {
    switch (scope.type) {
        case "settings":
            return ROUTE.settings.backupSnapshots.restore.$route(snapshotId);
        case "project":
            return ROUTE.projects.single.providerConfiguration.backupSnapshots.restore.$route(
                scope.projectId,
                snapshotId,
            );
        case "app":
            return ROUTE.projects.single.apps.single.configuration.backupSnapshots.restore.$route(
                scope.projectId,
                scope.env,
                scope.appId,
                snapshotId,
            );
    }
}
