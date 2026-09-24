import type { SettingsBaseEntity } from "~/settings/domain";

/** What the server sends in place of a stored secret. Sending it back keeps the secret. */
export const LOGGING_MASKED_SECRET = "********";

export type HivePaaSLoggingEndpoint = {
    url: string;
    username?: string;
    password?: string;
    bearerToken?: string;
    headers?: Record<string, string>;
    tlsSkipVerify?: boolean;
};

export type HivePaaSLoggingVictoriaLogs = {
    /** The volume decides both where logs are stored and which node the backend runs on. */
    volume?: { id: string; name?: string } | null;
    /** timeutil.Duration text: the server writes days as "30d", and accepts w/d/h/m/s. */
    retention: string;
    maxDiskUsagePercent?: number;
    /**
     * Cores, read off the backend's service - the same place the app's own
     * resource screen reads. Absent means no cap.
     */
    cpuLimit?: number;
    /** A size carrying its unit, such as "1gb", read off the service too. */
    memoryLimit?: string;
};

/** One app of the logging stack: an ordinary app in the hidden hivepaas project. */
export type HivePaaSLoggingApp = {
    appId: string;
    projectId: string;
    projectEnv: string;
    /** Containers: one for the backend, one per node for the collector. */
    runningTasks: number;
    desiredTasks: number;
};

export type HivePaaSLoggingStatus = {
    collectorReady: boolean;
    backendReady: boolean;
    /** The apps HivePaaS runs, absent when it runs none. */
    backend?: HivePaaSLoggingApp;
    collector?: HivePaaSLoggingApp;
};

export interface HivePaaSLoggingSettings extends SettingsBaseEntity {
    enabled: boolean;
    sources: { apps: boolean; hivepaas: boolean; traefikAccess: boolean; nodes: boolean };
    collector: { type: string; managed: boolean };
    backend: {
        type: string;
        managed: boolean;
        ingest?: HivePaaSLoggingEndpoint | null;
        query?: HivePaaSLoggingEndpoint | null;
        victoriaLogs?: HivePaaSLoggingVictoriaLogs | null;
    };
    forwards: { name: string; format?: string; endpoint: HivePaaSLoggingEndpoint }[];
    secretMasked?: boolean;
    loggingStatus?: HivePaaSLoggingStatus;
}
