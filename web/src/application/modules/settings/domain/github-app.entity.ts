import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingGithubApp extends SettingsBaseEntity {
    clientId: string;
    clientSecret: string;
    organization: string;
    callbackURL?: string;
    webhookURL?: string;
    webhookSecret?: string;
    appId: number;
    installationId: number;
    privateKey: string;
    ssoEnabled: boolean;
    secretMasked?: boolean;
}
