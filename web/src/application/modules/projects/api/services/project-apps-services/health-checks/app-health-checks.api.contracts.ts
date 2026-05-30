import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    AppHealthCheck,
    AppHealthCheckGRPC,
    AppHealthCheckNotification,
    AppHealthCheckREST,
} from "~/projects/domain";
import type { EAppHealthCheckType } from "~/projects/module-shared/enums";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppHealthChecks_RefObject_Payload = {
    id: string;
};

export type AppHealthChecks_REST_Payload = AppHealthCheckREST;
export type AppHealthChecks_GRPC_Payload = AppHealthCheckGRPC;

export type AppHealthChecks_Notification_Payload = Omit<AppHealthCheckNotification, "success" | "failure"> & {
    success?: AppHealthChecks_RefObject_Payload;
    failure?: AppHealthChecks_RefObject_Payload;
};

export type AppHealthChecks_Upsert_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    healthcheckType: EAppHealthCheckType;
    interval: string;
    maxRetry: number;
    retryDelay: string;
    timeout: string;
    saveResultTasks: boolean;
    rest: AppHealthChecks_REST_Payload | null;
    grpc: AppHealthChecks_GRPC_Payload | null;
    notification: AppHealthChecks_Notification_Payload;
};

export type AppHealthChecks_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppHealthChecks_FindManyPaginated_Res = ApiResponsePaginated<AppHealthCheck>;

export type AppHealthChecks_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    healthCheckID: string;
}>;

export type AppHealthChecks_FindOneById_Res = ApiResponseBase<AppHealthCheck>;

export type AppHealthChecks_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppHealthChecks_Upsert_Payload;
}>;

export type AppHealthChecks_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type AppHealthChecks_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    healthCheckID: string;
    payload: AppHealthChecks_Upsert_Payload & {
        updateVer: number;
    };
}>;

export type AppHealthChecks_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppHealthChecks_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    healthCheckID: string;
    payload: {
        updateVer: number;
        status: ESettingStatus;
        expireAt: Date | null;
        availableInProjects: boolean;
        default: boolean;
    };
}>;

export type AppHealthChecks_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type AppHealthChecks_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    healthCheckID: string;
}>;

export type AppHealthChecks_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;
