export const AuditLogResult = {
    Allowed: "allowed",
    Denied: "denied",
} as const;

export type AuditLogResult = (typeof AuditLogResult)[keyof typeof AuditLogResult];

export const AuditLogType = {
    SecretReveal: "secret-reveal",
    ApiKeyCreate: "api-key-create",
    ApiKeyRevoke: "api-key-revoke",
    SecuritySettingsUpdate: "security-settings-update",
    RoutingChangeConfirm: "routing-change-confirm",
    RoutingChangeRevert: "routing-change-revert",
    SettingCreate: "setting-create",
    SettingUpdate: "setting-update",
    SettingStatusUpdate: "setting-status-update",
    SettingDelete: "setting-delete",
    ProjectUpdate: "project-update",
    AppUpdate: "app-update",
} as const;

export type AuditLogType = (typeof AuditLogType)[keyof typeof AuditLogType];

export const AuditLogSource = {
    ApiGet: "api-get",
    ApiCreate: "api-create",
    CapabilityRevoked: "capability-revoked",
    ApiUpdate: "api-update",
    ApiDelete: "api-delete",
} as const;

export type AuditLogSource = (typeof AuditLogSource)[keyof typeof AuditLogSource];

export interface AuditLogActor {
    id: string;
    type: string;
    name?: string;
    loggedName?: string;
    photo?: string;
}

export interface AuditLogResource {
    id: string;
    type: string;
    name?: string;
    loggedName?: string;
    photo?: string;
}

export interface AuditLogScopeProject {
    id: string;
    name: string;
    key?: string;
    photo?: string;
    status?: string;
}

export interface AuditLogScopeProjectEnv {
    id: string;
    name: string;
    color?: string;
    status?: string;
}

export interface AuditLogScopeApp {
    id: string;
    name: string;
    key?: string;
    photo?: string;
    status?: string;
    env?: string;
}

export interface AuditLogScopeUser {
    id: string;
    username: string;
    email?: string;
    fullName?: string;
    photo?: string;
    role?: string;
}

export interface AuditLog {
    id: string;
    type: AuditLogType;
    source?: AuditLogSource;
    result: AuditLogResult;
    scopeProject?: AuditLogScopeProject;
    scopeApp?: AuditLogScopeApp;
    scopeUser?: AuditLogScopeUser;
    actor?: AuditLogActor;
    resource?: AuditLogResource;
    viaApiKey?: boolean;
    sessionUid?: string;
    clientIp?: string;
    remoteAddr?: string;
    userAgent?: string;
    requestId?: string;
    detail?: string;
    hasDetail?: boolean;
    createdAt: Date;
}

export type AuditLogScope =
    | { type: "global" }
    | { type: "project"; projectID: string }
    | { type: "project-env"; projectID: string; projectEnvID: string }
    | { type: "app"; projectID: string; projectEnvID: string; appID: string };
