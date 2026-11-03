import type { EProjectAppStatus } from "~/projects/module-shared/enums";

export type AppCloneSettingRef = {
    id: string;
    name: string;
};

export type AppCloneRoutingDomainSettings = {
    sourceDomain: string;
    targetDomain: string;
    sourceSslCert?: AppCloneSettingRef | null;
    targetSslCert?: AppCloneSettingRef | null;
};
export type AppCloneHttpDomainSettings = AppCloneRoutingDomainSettings;

export type AppCloneEventNotification = {
    successUseDefault: boolean;
    success?: AppCloneSettingRef | null;
    failureUseDefault: boolean;
    failure?: AppCloneSettingRef | null;
};

export interface AppCloneSettings {
    targetName: string;
    targetEnv: string;
    targetStatus: EProjectAppStatus;
    targetReplicas: number;
    cloneDeploymentSettings: boolean;
    cloneRoutingSettings: boolean;
    cloneRoutingDomains: AppCloneRoutingDomainSettings[];

    cloneVolumes: boolean;
    cloneVolumeData: boolean;
    liveVolumeClone: boolean;
    includedVolumes: string[];
    excludedVolumes: string[];
    cloneEnvVars: boolean;
    cloneSecrets: boolean;
    cloneConfigFiles: boolean;
    clonePeriodicJobs: boolean;
    cloneSchedJobs: boolean;
    commandPipes: AppCloneSettingRef[];
    notification?: AppCloneEventNotification | null;
    updateVer: number;
}
