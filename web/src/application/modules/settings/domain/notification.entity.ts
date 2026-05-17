import { type ESettingType } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface NotificationViaEmail {
    enabled: boolean;
    sender?: SettingsBaseEntity;
    toProjectMembers: boolean;
    toProjectOwners: boolean;
    toAllAdmins: boolean;
    toAddresses: string[];
}

export interface NotificationViaSlack {
    enabled: boolean;
    webhook?: SettingsBaseEntity;
}

export interface NotificationViaDiscord {
    enabled: boolean;
    webhook?: SettingsBaseEntity;
}

export interface SettingNotification extends SettingsBaseEntity {
    type: typeof ESettingType.Notification;
    viaEmail?: NotificationViaEmail;
    viaSlack?: NotificationViaSlack;
    viaDiscord?: NotificationViaDiscord;
    minSendInterval: string; // duration in string format from BE
    inherited?: boolean;
}
