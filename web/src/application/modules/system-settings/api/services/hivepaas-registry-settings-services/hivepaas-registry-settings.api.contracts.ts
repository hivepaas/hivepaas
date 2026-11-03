import type {
    HivePaaSRegistryDomainProbe,
    HivePaaSRegistryPushCheck,
    HivePaaSRegistrySettings,
    HivePaaSRegistryStatus,
    HivePaaSRegistryStorageType,
} from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSRegistrySettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSRegistrySettings_FindOne_Res = ApiResponseBase<{
    settings: HivePaaSRegistrySettings;
    registryStatus: HivePaaSRegistryStatus;
}>;

export type HivePaaSRegistrySettings_UpdateOnePayload = {
    updateVer: number;
    enabled: boolean;
    domain: string;
    storage: {
        type: HivePaaSRegistryStorageType;
        volume?: { id: string } | null;
        cloudStorage?: { id: string } | null;
    };
    cleanup: { enabled: boolean; keepLast: number; keepDays: number };
    dashboardEnabled: boolean;
    cpuLimit: number;
    memoryLimit: string;
    /**
     * Switching the registry off takes its app down, and the app has no screen of
     * its own to be removed from, so the confirmation is sent with the save.
     */
    removeApp?: boolean;
    /** Delete the images with the app: the registry's directory in the volume. */
    removeStorage?: boolean;
};

export type HivePaaSRegistrySettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSRegistrySettings_UpdateOnePayload;
}>;
export type HivePaaSRegistrySettings_UpdateOne_Res = ApiResponseBase<{
    /** The registry's app was taken down by this save. */
    removedApp: boolean;
    /** The registry account was left in place because an app still names it. */
    credentialKept: boolean;
}>;

export type HivePaaSRegistrySettings_ProbeDomain_Req = ApiRequestBase<{ domain: string }>;
export type HivePaaSRegistrySettings_ProbeDomain_Res = ApiResponseBase<HivePaaSRegistryDomainProbe>;

/** Bytes is optional: the server's default is above the limit a proxy on a free plan imposes. */
export type HivePaaSRegistrySettings_CheckPush_Req = ApiRequestBase<{ bytes?: number }>;
export type HivePaaSRegistrySettings_CheckPush_Res = ApiResponseBase<HivePaaSRegistryPushCheck>;

export type HivePaaSRegistrySettings_RotateCredential_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSRegistrySettings_RotateCredential_Res = ApiResponseBase<{ graceEndsAt: string }>;
