/**
 * A managed app: one HivePaaS's engine for it looks after once it is created - a
 * PostgreSQL database it archives, backs up, tests and restores. The engine owns
 * part of the app (its locks), and keeps its own state (its status).
 */

export const MANAGED_RESTORE_TEST_MODE = {
    container: "container",
    verifier: "verifier",
} as const;
export type ManagedRestoreTestMode = (typeof MANAGED_RESTORE_TEST_MODE)[keyof typeof MANAGED_RESTORE_TEST_MODE];

export const MANAGED_RESTORE_TARGET = {
    latest: "latest",
    time: "time",
    point: "point",
    backup: "backup",
} as const;
export type ManagedRestoreTargetKind = (typeof MANAGED_RESTORE_TARGET)[keyof typeof MANAGED_RESTORE_TARGET];

/** Where the archive is from the database: object storage, a volume any node mounts, or the database's own node. */
export type ManagedPlacement = "off-site" | "shared" | "same-node" | "";

export interface ManagedObjectRef {
    id: string;
}

/** What the engine owns of the app: the settings screens show it read-only. */
export interface AppManagedLocks {
    image: string;
    imageLine: string;
    command: string;
    mounts: string[];
    envKeys: string[];
    settings: string[];
}

export interface AppManagedStorage {
    backupRepo?: ManagedObjectRef;
    cloudStorage?: ManagedObjectRef;
    keyAuth?: ManagedObjectRef;
    prefix: string;
    placement: ManagedPlacement;
}

export interface AppManagedRestoreTestConfig {
    mode: ManagedRestoreTestMode;
    verifierApp?: ManagedObjectRef;
    nodeId: string;
    keyAuth?: ManagedObjectRef;
    checkQuery: string;
}

export interface AppManagedRestoreTarget {
    kind: ManagedRestoreTargetKind;
    time?: Date;
    point: string;
    lsn: string;
    backup: string;
}

/** Where a database restored into a new app came from; it is restoring until readyAt. */
export interface AppManagedRestoreOrigin {
    sourceApp?: ManagedObjectRef;
    target?: AppManagedRestoreTarget;
    readyAt?: Date;
    /** When the source's consumers were last switched to it. */
    switchedAt?: Date;
}

/** An app of the environment as a switch names it. */
export interface AppManagedSwitchApp {
    id: string;
    name: string;
    key: string;
}

/** An app that uses the database a restored one was restored from, and how. */
export interface AppManagedSwitchConsumer {
    app: AppManagedSwitchApp;
    /** Its variables that name the source's. */
    envVars: string[];
    /** Its previews clone the source. */
    previewClone: boolean;
    /** Its deployments mark a restore point on the source. */
    deployRestorePoint: boolean;
}

/** What switching to a restored database would repoint: its source, which keeps running, and its consumers. */
export interface AppManagedSwitch {
    source: AppManagedSwitchApp;
    consumers: AppManagedSwitchConsumer[];
    switchedAt?: Date;
}

/** The pg_dump into a backup repository: none while dumps are off; the schedule is kept either way, empty for the engine's default. */
export interface AppManagedDump {
    repo?: ManagedObjectRef;
    schedule: string;
}

/** A replica of a primary: a logical child on a volume of its own. */
export interface AppManagedMember {
    app?: ManagedObjectRef;
    key: string;
    role: string;
    volume?: ManagedObjectRef;
    /** The node its volume is pinned to: an id, or a label. */
    node: string;
    at: Date;
}

/** A replica as the primary's last health check saw it. */
export interface AppManagedMemberStatus {
    app?: ManagedObjectRef;
    key: string;
    role: string;
    node: string;
    /** pg_stat_replication's state, "disconnected", or "unknown" before a check has looked. */
    state: string;
    lagBytes: number;
    /** Seconds behind; none while nothing is written. */
    lagSeconds?: number;
    checkedAt?: Date;
}

export interface AppManagedSettings {
    updateVer: number;
    engine: string;
    engineApi: number;
    storage?: AppManagedStorage;
    archiveTimeout: string;
    retainFull: number;
    baseBackupSchedule: string;
    restoreTestSchedule: string;
    restoreTest?: AppManagedRestoreTestConfig;
    dump: AppManagedDump;
    locks: AppManagedLocks;
    jobs?: {
        nightly?: ManagedObjectRef;
        health?: ManagedObjectRef;
        restoreTest?: ManagedObjectRef;
        dump?: ManagedObjectRef;
    };
    restore?: AppManagedRestoreOrigin;
    /** The database restored from this one that its consumers were switched to: this one's jobs are paused since. */
    switchedTo?: { app?: ManagedObjectRef; at: Date };
    members: AppManagedMember[];
    /** The replica the base backups are taken on; none for this database itself. */
    backupFrom?: ManagedObjectRef;
    /** The member promoted to the cluster's primary; none while this database is. */
    writer?: ManagedObjectRef;
    /** When this database was fenced for a promotion: stopped, and kept from starting its server. */
    fencedAt?: Date;
    /** The primary this database is a replica of. */
    memberOf?: { app?: ManagedObjectRef; role: string };
}

export const MANAGED_ARCHIVING_STATE = {
    ok: "ok",
    failing: "failing",
    unknown: "unknown",
} as const;
export type ManagedArchivingState = (typeof MANAGED_ARCHIVING_STATE)[keyof typeof MANAGED_ARCHIVING_STATE];

export interface AppManagedRestorePoint {
    name: string;
    lsn: string;
    at: Date;
    by?: ManagedObjectRef;
    note: string;
    keepUntil?: Date;
    /** Retention is held back for it now. */
    kept: boolean;
    /** Until when it can be restored, as retention's next runs go; none when no base backup kept reaches it. */
    restorableUntil?: Date;
    /** What made it: a person, or a deployment of an app that asks for one before each deploy. */
    origin: "person" | "deploy";
}

export interface AppManagedStatus {
    engine: string;
    engineApi: number;
    archiving: {
        state: ManagedArchivingState;
        checkedAt?: Date;
        lastArchivedAt?: Date;
        failedCount: number;
        pgWalBytes: number;
        stale: boolean;
    };
    /** What a restore may target; none before the first base backup. */
    restoreWindow?: { from: Date; to: Date };
    lastBackup?: { succeeded: boolean; endedAt: Date; name: string; at?: Date; bytes: number };
    /** The base backups kept, oldest first. */
    backups: { name: string; at: Date }[];
    lastRestoreTest?: { passed: boolean; endedAt: Date; durationSeconds: number; restoredTo: string };
    /** The dump's last run, and the snapshot it took; none while dumps are off or before the first. */
    lastDump?: { succeeded: boolean; endedAt: Date; snapshotId: string; bytes: number };
    placement: ManagedPlacement;
    /**
     * Changes worth making: "verifier", to run the restore test in a verifier app; "dump-apart", to dump at
     * another time than the base backup, as both read the whole database.
     */
    suggestions: string[];
    restorePoints: AppManagedRestorePoint[];
    /** What the kept points make retention keep beyond its own. */
    heldBack?: { backups: number; bytes: number };
    members: AppManagedMemberStatus[];
}
