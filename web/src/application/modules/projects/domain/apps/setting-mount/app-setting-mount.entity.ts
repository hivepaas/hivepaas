import type { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

/** Why nothing of an entry is mounted; empty when something is. */
export const SETTING_MOUNT_REASONS = {
    "entry-disabled": "The entry is disabled.",
    "key-invalid": "The name is not one an entry may have.",
    "source-unavailable": "The source setting is missing, disabled, or not visible from this app.",
    "source-incomplete": "The source setting has no value yet, such as a certificate not obtained.",
    "paths-taken": "Every path is used by another file of the app.",
} as const;

export type SettingMountReason = keyof typeof SETTING_MOUNT_REASONS;

export interface AppSettingMountFile {
    part: string;
    path: string;
    uid: string;
    gid: string;
    mode: string;
    /** Stored as a Docker secret rather than a config. */
    secret: boolean;
    /** Mounting it takes the Reveal Secrets permission. */
    gated: boolean;
}

export interface AppSettingMountState {
    /** Empty when mounted. */
    reason: string;
    mounted: string[];
}

export interface AppSettingMountSourceRef {
    id: string;
    name: string;
    type: string;
    status: string;
}

export interface AppSettingMount {
    id: string;
    /** The entry's key. */
    name: string;
    status: OpenApiConstant<EProjectSecretStatus>;
    /** Previews and clones of the app get the entry. */
    inheritable: boolean;
    updateVer: number;
    source: AppSettingMountSourceRef;
    files: AppSettingMountFile[];
    /** Null when the backend could not read it. */
    state: AppSettingMountState | null;
    createdAt: Date;
    updatedAt: Date | null;
}

export interface AppSettingMountSourcePart {
    name: string;
    required: boolean;
    secret: boolean;
    gated: boolean;
}

export interface AppSettingMountSource {
    type: string;
    parts: AppSettingMountSourcePart[];
}

export interface AppSettingMountSources {
    sources: AppSettingMountSource[];
    /** Whether the caller may mount a gated part: a private key, a password. */
    mayMountSensitive: boolean;
}
