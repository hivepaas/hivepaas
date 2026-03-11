import { type ESettingType } from "@application/shared/enums";

import type { SettingsBaseEntity } from "../settings.base.entity";

export interface NotificationViaEmailEntity {
    sender?: SettingsBaseEntity;
    toProjectMembers: boolean;
    toProjectOwners: boolean;
    toAllAdmins: boolean;
    toAddresses: string[];
}

export interface NotificationViaSlackEntity {
    webhook?: SettingsBaseEntity;
}

export interface NotificationViaDiscordEntity {
    webhook?: SettingsBaseEntity;
}

export interface NotificationEntity extends SettingsBaseEntity {
    type: typeof ESettingType.Notification;
    viaEmail?: NotificationViaEmailEntity;
    viaSlack?: NotificationViaSlackEntity;
    viaDiscord?: NotificationViaDiscordEntity;
    minSendInterval: string; // duration in string format from BE
}
