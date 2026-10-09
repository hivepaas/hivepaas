import type {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import type { EUserRole } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

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
    /** The change the deployment was made for, as its trigger gave it: pr-<number> for a pull request. */
    changeId: string;
}

export interface AppDeploymentOutput {
    commitHash: string;
    commitHashShort: string;
    commitURL: string;
    commitTitle: string;
    commitMessage: string;
    commitAuthor: string;
    error: string;
    imageTags: string[];
}

export type AppDeploymentRepoSnapshot = {
    repoUrl: string;
    repoRef: string;
};

type AppDeploymentRepoSettingsSnapshot = {
    activeMethod: typeof EAppDeploymentMethod.Repo;
    repoSource: AppDeploymentRepoSnapshot;
};

type AppDeploymentImageSettingsSnapshot = {
    activeMethod: typeof EAppDeploymentMethod.Image;
};

/** A function's deployment: its repository, or null for inline code. */
type AppDeploymentFunctionSettingsSnapshot = {
    activeMethod: typeof EAppDeploymentMethod.Function;
    repoSource: AppDeploymentRepoSnapshot | null;
};

export type AppDeploymentSettingsSnapshot =
    AppDeploymentRepoSettingsSnapshot | AppDeploymentImageSettingsSnapshot | AppDeploymentFunctionSettingsSnapshot;

export interface AppDeployment {
    id: string;
    status: OpenApiConstant<EAppDeploymentStatus>;
    updateVer: number;
    settings: AppDeploymentSettingsSnapshot;
    trigger: AppDeploymentTrigger | null;
    output: AppDeploymentOutput | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date | null;
}

/** The app's deployment that has not ended: the one running, or else the next to run. */
export interface AppActiveDeployment {
    id: string;
    status: typeof EAppDeploymentStatus.InProgress | typeof EAppDeploymentStatus.NotStarted;
}
