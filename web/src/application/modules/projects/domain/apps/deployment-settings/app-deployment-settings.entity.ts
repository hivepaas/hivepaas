import { type EAppDeploymentMethod, type EBuildTool, type ERepoType } from "~/projects/module-shared/enums";
import { type SettingsBaseEntity } from "~/settings/domain";

export type RepoMethod = BaseDeploymentSettings & {
    activeMethod: typeof EAppDeploymentMethod.Repo;
    repoSource: {
        buildTool: EBuildTool;
        repoType: ERepoType;
        repoUrl: string;
        repoRef: string;
        credentials: SettingsBaseEntity;
        dockerfilePath: string;
        imageName: string;
        imageTags: string;
        pushToRegistry: SettingsBaseEntity;
    };
};

export type ImageMethod = BaseDeploymentSettings & {
    activeMethod: typeof EAppDeploymentMethod.Image;
    image: string;
    registryAuth: SettingsBaseEntity;
};

export type BaseDeploymentSettings = {
    command?: string;
    workingDir?: string;
    preDeploymentCommand?: string;
    postDeploymentCommand?: string;

    notification?: {
        success: SettingsBaseEntity;
        failure: SettingsBaseEntity;
    };

    updateVer: number;
};

export type AppDeploymentSettings = RepoMethod | ImageMethod;
