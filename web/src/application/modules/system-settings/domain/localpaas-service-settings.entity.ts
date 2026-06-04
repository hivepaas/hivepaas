import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingType } from "@application/shared/enums";

export interface LocalPaaSAppSettings {
    replicas: number;
}

export interface LocalPaaSWorkerSettings {
    replicas: number;
    concurrency: number;
    runWorkerInMainApp: boolean;
}

export interface LocalPaaSTaskSettings {
    taskCheckInterval: string;
    taskCreateInterval: string;
}

export interface LocalPaaSHealthcheckSettings {
    baseInterval: string;
}

export interface LocalPaaSServiceSettings extends SettingsBaseEntity {
    type: typeof ESettingType.LocalPaaSService;
    appSettings: LocalPaaSAppSettings;
    workerSettings: LocalPaaSWorkerSettings;
    taskSettings: LocalPaaSTaskSettings;
    healthcheckSettings: LocalPaaSHealthcheckSettings;
}
