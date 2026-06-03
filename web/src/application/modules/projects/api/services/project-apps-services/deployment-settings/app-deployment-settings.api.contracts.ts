import { type EAppDeploymentMethod, type EBuildTool, type ERepoType } from "~/projects/module-shared/enums";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

import { type AppDeploymentSettings } from "../../../../domain/apps/deployment-settings";

export type AppDeploymentSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppDeploymentSettings_FindOne_Res = ApiResponseBase<AppDeploymentSettings>;

type AppDeploymentSettings_UpdateBasePayload = {
    command: string;
    workingDir: string;
    preDeploymentCommand: string;
    postDeploymentCommand: string;
    notification: {
        successUseDefault: boolean;
        success?: {
            id: string;
        };
        failureUseDefault: boolean;
        failure?: {
            id: string;
        };
    };
};

type AppDeploymentSettings_UpdateImagePayload = AppDeploymentSettings_UpdateBasePayload & {
    activeMethod: typeof EAppDeploymentMethod.Image;
    imageSource: {
        image: string;
        registryAuth: {
            id: string;
        };
    };
};

type AppDeploymentSettings_UpdateRepoPayload = AppDeploymentSettings_UpdateBasePayload & {
    activeMethod: typeof EAppDeploymentMethod.Repo;
    repoSource: {
        buildTool: EBuildTool;
        repoType: ERepoType;
        repoUrl: string;
        repoRef: string;
        commitHash: string;
        repoOptions: {
            gitSubmodulesEnabled: boolean;
            gitLfsEnabled: boolean;
        };
        credentials: {
            id: string;
        };
        dockerfilePath?: string;
        imageName: string;
        imageTags?: string;
        pushToRegistry: {
            id: string;
        };
    };
};

export type AppDeploymentSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    updateVer: number;
    payload: AppDeploymentSettings_UpdateImagePayload | AppDeploymentSettings_UpdateRepoPayload;
}>;
export type AppDeploymentSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
