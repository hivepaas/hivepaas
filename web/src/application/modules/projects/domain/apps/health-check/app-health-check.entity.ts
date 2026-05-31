import type {
    EAppHealthCheckGrpcStatus,
    EAppHealthCheckGrpcVersion,
    EAppHealthCheckRestMethod,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";
import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

export interface AppHealthCheckNamedRef {
    id: string;
    name: string;
}

export interface AppHealthCheckRestReturnText {
    exact: string;
    regex: string;
}

export interface AppHealthCheckRestReturnJSON {
    exact: string;
    contain: string;
}

export interface AppHealthCheckREST {
    url: string;
    method: EAppHealthCheckRestMethod;
    contentType: string;
    body: string;
    returnCode: string;
    returnText: AppHealthCheckRestReturnText | null;
    returnJSON: AppHealthCheckRestReturnJSON | null;
}

export interface AppHealthCheckGRPC {
    version: EAppHealthCheckGrpcVersion;
    addr: string;
    service: string;
    returnStatus: EAppHealthCheckGrpcStatus;
}

export interface AppHealthCheckNotification {
    successUseDefault: boolean;
    success?: AppHealthCheckNamedRef;
    failureUseDefault: boolean;
    failure?: AppHealthCheckNamedRef;
    minSendInterval: string;
}

export interface AppHealthCheck {
    id: string;
    type: ESettingType;
    name: string;
    kind?: string;
    status: ESettingStatus;
    inherited: boolean;
    availableInProjects: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date;
    expireAt: Date | null;

    healthcheckType: EAppHealthCheckType;
    interval: string;
    maxRetry: number;
    retryDelay: string;
    timeout: string;
    saveResultTasks: boolean;
    rest: AppHealthCheckREST | null;
    grpc: AppHealthCheckGRPC | null;
    notification: AppHealthCheckNotification | null;
}

export type AppHealthCheckNotificationRef = Pick<SettingsBaseEntity, "id" | "name">;
