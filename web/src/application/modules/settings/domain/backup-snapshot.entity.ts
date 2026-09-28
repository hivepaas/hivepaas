import type { AppScheduledJobCommand } from "~/projects/domain";

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
    job?: BackupSnapshotJob;
    /** The task of the run that took it. */
    runId: string;
}

/**
 * The job that took a snapshot, and what a restore needs of it: the file a command's snapshot holds and the
 * command that loads it, or the volume a volume's was read from.
 */
export interface BackupSnapshotJob {
    id: string;
    name: string;
    deleted: boolean;
    fileName: string;
    restoreCommand?: AppScheduledJobCommand;
    sourceVolumeId: string;
    sourceVolumeSubpath: string;
}

/** A file or a directory a snapshot holds; a directory's size is all it holds. */
export interface BackupSnapshotEntry {
    name: string;
    dir: boolean;
    sizeBytes: number;
}

/** How a restore writes a snapshot into a directory. */
export const BACKUP_RESTORE_MODE = {
    /** The directory is moved aside, and the snapshot restored into an empty one. */
    Replace: "replace",
    /** The snapshot's files are written over what is there. */
    Overwrite: "overwrite",
} as const;

export type BackupRestoreMode = (typeof BACKUP_RESTORE_MODE)[keyof typeof BACKUP_RESTORE_MODE];
