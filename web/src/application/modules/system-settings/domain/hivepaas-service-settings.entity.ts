import type { SettingsBaseEntity } from "~/settings/domain";

import type { SettingsPendingChange } from "./settings-probation.entity";

export interface HivePaaSAppSettings {
    replicas: number;
}

export interface HivePaaSWorkerSettings {
    replicas: number;
    concurrency: number;
    runWorkerInMainApp: boolean;
}

export interface HivePaaSTaskSettings {
    taskCheckInterval: string;
    taskCreateInterval: string;
}

export interface HivePaaSPeriodicSettings {
    baseInterval: string;
    batchSize: number;
}

export interface HivePaaSProxySettings {
    proxyProvider: string;
    trustedIPs: string[];
    proxyHops: number;
}

export interface HivePaaSServiceSettings extends SettingsBaseEntity {
    appSettings: HivePaaSAppSettings;
    workerSettings: HivePaaSWorkerSettings;
    taskSettings: HivePaaSTaskSettings;
    periodicSettings: HivePaaSPeriodicSettings;
    proxySettings: HivePaaSProxySettings;

    /**
     * Set while a proxy settings change is on trial. Only the proxy fields go on
     * trial - they are the ones that decide how a client address is read, and so
     * whether the IP allowlists still admit the caller.
     */
    pendingChange: SettingsPendingChange | null;
}
