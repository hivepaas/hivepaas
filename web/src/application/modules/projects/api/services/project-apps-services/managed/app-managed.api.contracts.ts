import type {
    AppManagedRestorePoint,
    AppManagedSettings,
    AppManagedStatus,
    AppManagedSwitch,
    ManagedRestoreTargetKind,
    ManagedRestoreTestMode,
} from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

type AppScope = { projectID: string; env: string; appID: string };

export type AppManaged_FindSettings_Req = ApiRequestBase<AppScope>;
export type AppManaged_FindSettings_Res = ApiResponseBase<AppManagedSettings>;

export type AppManaged_FindStatus_Req = ApiRequestBase<AppScope>;
export type AppManaged_FindStatus_Res = ApiResponseBase<AppManagedStatus>;

/** A field left out is left as it is; an empty query clears it. */
export type AppManagedSettings_UpdatePayload = {
    updateVer: number;
    retainFull?: number;
    baseBackupSchedule?: string;
    restoreTestSchedule?: string;
    restoreTestQuery?: string;
    /** A repository turns dumps on, an empty id turns them off. */
    dumpRepo?: { id: string };
    dumpSchedule?: string;
    /** A replica takes the base backups; an empty id has the database take them. */
    backupFrom?: { id: string };
    restoreTest?: {
        mode: ManagedRestoreTestMode;
        volume?: { id: string };
        keyAuth?: { id: string };
    };
};

export type AppManaged_UpdateSettings_Req = ApiRequestBase<AppScope & { payload: AppManagedSettings_UpdatePayload }>;
export type AppManaged_UpdateSettings_Res = ApiResponseBase<{ type: "success" }>;

export type AppManaged_RefreshCredentials_Req = ApiRequestBase<AppScope>;
export type AppManaged_RefreshCredentials_Res = ApiResponseBase<{ changed: boolean }>;

export type AppManaged_Unmanage_Req = ApiRequestBase<AppScope>;
export type AppManaged_Unmanage_Res = ApiResponseBase<{ type: "success" }>;

export type AppManagedRestorePayload = {
    target: { kind: ManagedRestoreTargetKind; time?: string; point?: string; backup?: string };
    name?: string;
    dataVolume: { id: string };
};

export type AppManaged_Restore_Req = ApiRequestBase<AppScope & { payload: AppManagedRestorePayload }>;
export type AppManaged_Restore_Res = ApiResponseBase<{ app: { id: string }; task: { id: string } }>;

export type AppManagedRestorePoint_CreatePayload = { name: string; note?: string; keepUntil?: string };

export type AppManaged_CreateRestorePoint_Req = ApiRequestBase<
    AppScope & { payload: AppManagedRestorePoint_CreatePayload }
>;
export type AppManaged_CreateRestorePoint_Res = ApiResponseBase<AppManagedRestorePoint>;

/** Both as given: a null keepUntil keeps the point no longer. */
export type AppManagedRestorePoint_UpdatePayload = { keepUntil: string | null; note: string };

export type AppManaged_UpdateRestorePoint_Req = ApiRequestBase<
    AppScope & { name: string; payload: AppManagedRestorePoint_UpdatePayload }
>;
export type AppManaged_UpdateRestorePoint_Res = ApiResponseBase<{ type: "success" }>;

export type AppManaged_DeleteRestorePoint_Req = ApiRequestBase<AppScope & { name: string }>;
export type AppManaged_DeleteRestorePoint_Res = ApiResponseBase<{ type: "success" }>;

export type AppManaged_FindSwitch_Req = ApiRequestBase<AppScope>;
export type AppManaged_FindSwitch_Res = ApiResponseBase<AppManagedSwitch>;

/** The consumers chosen are switched; with redeploy, each whose variables changed is deployed too. */
export type AppManagedSwitchPayload = { consumers: { id: string }[]; redeploy: boolean };

export type AppManaged_Switch_Req = ApiRequestBase<AppScope & { payload: AppManagedSwitchPayload }>;
export type AppManaged_Switch_Res = ApiResponseBase<{
    switched: { id: string }[];
    deployments: { app: { id: string }; deployment: { id: string }; task: { id: string } }[];
}>;

export type AppManaged_AddReplica_Req = ApiRequestBase<AppScope & { payload: { dataVolume: { id: string } } }>;
export type AppManaged_AddReplica_Res = ApiResponseBase<{ app: { id: string }; task?: { id: string } }>;

export type AppManaged_RemoveReplica_Req = ApiRequestBase<AppScope & { memberID: string }>;
export type AppManaged_RemoveReplica_Res = ApiResponseBase<{ type: "success" }>;

export type AppManaged_Promote_Req = ApiRequestBase<AppScope & { payload: { member: { id: string } } }>;
export type AppManaged_Promote_Res = ApiResponseBase<{
    writer: { id: string };
    fenced: { id: string };
    restarted: { id: string }[];
}>;
