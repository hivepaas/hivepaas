import type { EEmailKind } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface SettingEmail extends SettingsBaseEntity {
    kind: EEmailKind;
    smtp?: EmailSMTP | null;
    http?: EmailHTTP | null;
    secretMasked?: boolean;
    inherited?: boolean;
}

export interface EmailSMTP {
    host: string;
    port: number;
    username: string;
    displayName: string;
    password: string;
    ssl: boolean;
}

export interface EmailHTTP {
    endpoint: string;
    method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
    contentType: string;
    headers?: Record<string, string> | null;
    fieldMapping?: EmailHTTPFieldMapping | null;
    username: string;
    displayName: string;
    password: string;
}

export interface EmailHTTPFieldMapping {
    fromAddress: string;
    fromName: string;
    toAddress: string;
    toAddresses: string;
    subject: string;
    content: string;
    password: string;
}
