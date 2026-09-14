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
    /** Directory inside the volume; empty means its root. */
    volumeSubpath?: string;
    /** timeutil.Duration text: the server writes days as "30d", and accepts w/d/h/m/s. */
    retention: string;
    maxDiskUsagePercent?: number;
    /** Cores. Absent means the backend may take whatever its node has. */
    cpuLimit?: number;
    /** A size carrying its unit, such as "1gb". Absent means no cap. */
    memoryLimit?: string;
};

export type HivePaaSLoggingStatus = {
    collectorReady: boolean;
    backendReady: boolean;
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
