export type ReleaseChannel = "stable" | "beta";

/** Where a published release stands against the one running. */
export type ReleaseRelation = "newer" | "same" | "older" | "";

export interface RunningRelease {
    appVersion: string;
    channel: ReleaseChannel;
    releaseDate: Date | null;
}

export interface PublishedRelease {
    appVersion: string;
    releaseDate: Date | null;
    notesUrl: string;
    /** Newer than the release running, so it may be moved to. */
    canUpdate: boolean;
    relation: ReleaseRelation;
}

export interface HivePaaSReleaseInfo {
    current: RunningRelease | null;
    stable: PublishedRelease | null;
    beta: PublishedRelease | null;
}

/** none, update, major, blocked or not-deployed. */
export type UpdateChange = "none" | "update" | "major" | "blocked" | "not-deployed";

export interface UpdateComponent {
    /** db, redis, traefik, victoria-logs, vlagent, registry, app or worker. */
    key: string;
    currentImage: string;
    targetImage: string;
    change: UpdateChange;
    /** The update's own words for what it would do. */
    reason: string;
    requiresBackup: boolean;
    interruptsTraffic: boolean;
}

export interface HivePaaSUpdatePlan {
    current: RunningRelease | null;
    target: { appVersion: string; channel: ReleaseChannel; releaseDate: Date | null; notesUrl: string };
    components: UpdateComponent[];
    /** The update would refuse to run. */
    blocked: boolean;
    /** The database backup cannot be skipped. */
    requiresBackup: boolean;
}
