import { type ESettingType } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface NotificationViaEmail {
    sender?: SettingsBaseEntity;
    toProjectMembers: boolean;
    toProjectOwners: boolean;
    toAllAdmins: boolean;
    toAddresses: string[];
}

export interface NotificationViaSlack {
    webhook?: SettingsBaseEntity;
}

export interface NotificationViaDiscord {
    webhook?: SettingsBaseEntity;
}

export interface SettingNotification extends SettingsBaseEntity {
    type: typeof ESettingType.Notification;
    viaEmail?: NotificationViaEmail;
    viaSlack?: NotificationViaSlack;
    viaDiscord?: NotificationViaDiscord;
    minSendInterval: string; // duration in string format from BE
}
