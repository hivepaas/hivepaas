/** A setting a snapshot names; its status is "missing" when it is deleted. */
export interface BackupSnapshotRepoRef {
    id: string;
    name: string;
    status: string;
    scope?: string;
}

/** The scope a Backup Snapshots view is of. */
export type BackupSnapshotScope =
    | { type: "settings" }
    | { type: "project"; projectId: string; env?: string }
    | { type: "app"; projectId: string; env: string; appId: string };

/** A snapshot a repository holds, with where it came from. */
export interface BackupSnapshot {
    /** The record's ID; snapshotId is the repository's. */
    id: string;
    snapshotId: string;
    shortId: string;
    time: Date;
    sizeBytes: number;
    description: string;
    paths: string[];
    hostname: string;
    tags: string[];
    /** command or volume; empty when not known. */
    source: string;
    repo: BackupSnapshotRepoRef;
    app?: { id: string; name: string; env: string; deleted: boolean };
    job?: { id: string; name: string; deleted: boolean };
    /** The task of the run that took it. */
    runId: string;
}
