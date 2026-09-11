export const AuditLogResult = {
    Allowed: "allowed",
    Denied: "denied",
} as const;

export type AuditLogResult = (typeof AuditLogResult)[keyof typeof AuditLogResult];

export const AuditLogType = {
    SecretReveal: "secret-reveal",
    ApiKeyCreate: "api-key-create",
    ApiKeyRevoke: "api-key-revoke",
    SettingCreate: "setting-create",
    SettingUpdate: "setting-update",
    SettingStatusUpdate: "setting-status-update",
    SettingDelete: "setting-delete",
    ProjectUpdate: "project-update",
    ProjectCreate: "project-create",
    ProjectDelete: "project-delete",
    AppUpdate: "app-update",
    AppCreate: "app-create",
    AppDelete: "app-delete",
    ProjectEnvDelete: "project-env-delete",
    ClusterUpdate: "cluster-update",
    TaskCancel: "task-cancel",
    UserLogin: "user-login",
    UserLogout: "user-logout",
    HivePaaSSecuritySettingsUpdate: "hivepaas-security-settings-update",
    HivePaaSSettingsUpdateConfirm: "hivepaas-settings-update-confirm",
    HivePaaSSettingsUpdateRevert: "hivepaas-settings-update-revert",
    HivePaaSSettingsUpdate: "hivepaas-settings-update",
    HivePaaSAction: "hivepaas-action",
} as const;

export type AuditLogType = (typeof AuditLogType)[keyof typeof AuditLogType];

export const AuditLogSource = {
    ApiGet: "api-get",
    ApiCreate: "api-create",
    CapabilityRevoked: "capability-revoked",
    ApiUpdate: "api-update",
    ApiDelete: "api-delete",
    ApiAction: "api-action",
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
    /**
     * Which part of the object was written, for the types that cover several
     * endpoints - the settings tab, the slice of a project. It used to live
     * inside `detail`, which the listing drops once it grows past ~100 chars, so
     * it went missing on exactly the entries that had the most to say.
     */
    section?: string;
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
