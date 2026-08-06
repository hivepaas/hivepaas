import type { AppCloneSettings } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppCloneSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppCloneSettings_FindOne_Res = ApiResponseBase<AppCloneSettings>;

export type AppCloneSettings_UpdateHttpDomainPayload = {
    sourceDomain: string;
    targetDomain: string;
    sourceSslCert: { id: string };
    targetSslCert: { id: string };
};

export type AppCloneSettings_UpdatePayload = {
    updateVer: number;
    targetName: string;
    targetEnv: string;
    targetStatus: AppCloneSettings["targetStatus"];
    targetReplicas: number;
    cloneDeploymentSettings: boolean;
    cloneHttpSettings: boolean;
    cloneHttpDomains: AppCloneSettings_UpdateHttpDomainPayload[];
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
    commandPipes: { id: string }[];
    notification: {
        successUseDefault: boolean;
        success?: { id: string };
        failureUseDefault: boolean;
        failure?: { id: string };
    };
};

export type AppCloneSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppCloneSettings_UpdatePayload;
}>;

export type AppCloneSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AppCloneSettings_Execute_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppCloneSettings_Execute_Res = ApiResponseBase<{ type: "success" }>;
