import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingNotification } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type Notification_RefObject_Payload = {
    id: string;
};

export type Notification_ViaEmail_Payload = {
    enabled: boolean;
    sender: Notification_RefObject_Payload;
    toProjectMembers: boolean;
    toProjectOwners: boolean;
    toAllAdmins: boolean;
    toAddresses: string[];
};

export type Notification_ViaSlack_Payload = {
    enabled: boolean;
    webhook: Notification_RefObject_Payload;
};

export type Notification_ViaDiscord_Payload = {
    enabled: boolean;
    webhook: Notification_RefObject_Payload;
};

export type Notification_ViaTelegram_Payload = {
    enabled: boolean;
    setting: Notification_RefObject_Payload;
};

export type Notification_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    viaEmail: Notification_ViaEmail_Payload | null;
    viaSlack: Notification_ViaSlack_Payload | null;
    viaDiscord: Notification_ViaDiscord_Payload | null;
    viaTelegram: Notification_ViaTelegram_Payload | null;
    minSendInterval: string;
};

/**
 * Notifications
 */
export type Notifications_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type Notifications_FindManyPaginated_Res = ApiResponsePaginated<SettingNotification>;

export type Notifications_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_FindOneById_Res = ApiResponseBase<SettingNotification>;

export type Notifications_CreateOne_Req = ApiRequestBase<{
    payload: Notification_CreateOne_Payload;
}>;
export type Notifications_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type Notifications_UpdateOne_Payload = Notification_CreateOne_Payload & {
    updateVer: number;
};

export type Notifications_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: Notifications_UpdateOne_Payload;
}>;
export type Notifications_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type Notifications_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type Notifications_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: Notifications_UpdateStatus_Payload;
}>;
export type Notifications_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type Notifications_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type Notifications_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
