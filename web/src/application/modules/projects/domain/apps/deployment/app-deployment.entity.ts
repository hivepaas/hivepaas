import type {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import type { EUserRole } from "@application/shared/enums";

export interface AppDeploymentSourceUser {
    id: string;
    username: string;
    email: string;
    fullName: string;
    photo: string | null;
    role: EUserRole;
}

export interface AppDeploymentTrigger {
    source: EAppDeploymentTriggerSource;
    sourceUser: AppDeploymentSourceUser | null;
}

export interface AppDeploymentOutput {
    commitHash: string;
    commitHashShort: string;
    commitURL: string;
    commitTitle: string;
    commitMessage: string;
    commitAuthor: string;
    imageTags: string[];
}

export interface AppDeploymentSettingsSnapshot {
    activeMethod: EAppDeploymentMethod;
}

export interface AppDeployment {
    id: string;
    status: EAppDeploymentStatus;
    updateVer: number;
    settings: AppDeploymentSettingsSnapshot;
    trigger: AppDeploymentTrigger | null;
    output: AppDeploymentOutput | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date | null;
}
