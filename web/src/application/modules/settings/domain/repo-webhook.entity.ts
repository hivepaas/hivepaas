import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingRepoWebhook extends SettingsBaseEntity {
    kind: string;
    secret: string;
    webhookURL: string;
    secretMasked?: boolean;
}
