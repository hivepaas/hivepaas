import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingType } from "@application/shared/enums";

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
}

export interface HivePaaSServiceSettings extends SettingsBaseEntity {
    type: typeof ESettingType.HivePaaSService;
    appSettings: HivePaaSAppSettings;
    workerSettings: HivePaaSWorkerSettings;
    taskSettings: HivePaaSTaskSettings;
    periodicSettings: HivePaaSPeriodicSettings;
    proxySettings: HivePaaSProxySettings;
}
