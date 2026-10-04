import {
    type EAppDeploymentMethod,
    type EBuildTool,
    type EDockerfileSource,
    type ERepoType,
} from "~/projects/module-shared/enums";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

import { type AppDeploymentSettings } from "../../../../domain/apps/deployment-settings";
import { type FunctionFile } from "../../../../domain/apps/function";

export type AppDeploymentSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppDeploymentSettings_FindOne_Res = ApiResponseBase<AppDeploymentSettings>;

type AppDeploymentSettings_UpdateBasePayload = {
    entrypoint: string;
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
        dockerfile: {
            source: EDockerfileSource;
            path: string;
            content?: string;
            scanPath?: string;
        };
        pushToRegistry: {
            id: string;
        };
        /** Whether a push to repoRef, received by a repo webhook, deploys the app. */
        autoDeploy: boolean;
    };
};

/**
 * A function's source as the API takes it: what FunctionSource says, the
 * settings it refers to by id.
 */
export type FunctionSourcePayload = {
    runtime: string;
    contract: string;
    entrypoint: { file: string; handler: string };
    code: {
        inline?: { files: FunctionFile[] };
        repo?: {
            repoType: string;
            repoUrl: string;
            repoRef: string;
            commitHash: string;
            credentials: { id: string };
            /** Whether a push to repoRef, received by a repo webhook, deploys the function. */
            autoDeploy: boolean;
        };
        dir: string;
    };
    systemPackages: string[];
    timeout: string;
    maxConcurrency: number;
    maxBodySize: string;
    pushToRegistry: { id: string };
};

type AppDeploymentSettings_UpdateFunctionPayload = AppDeploymentSettings_UpdateBasePayload & {
    activeMethod: typeof EAppDeploymentMethod.Function;
    functionSource: FunctionSourcePayload;
};

export type AppDeploymentSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    updateVer: number;
    payload:
        | AppDeploymentSettings_UpdateImagePayload
        | AppDeploymentSettings_UpdateRepoPayload
        | AppDeploymentSettings_UpdateFunctionPayload;
}>;
export type AppDeploymentSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AppDeploymentSettings_GetDockerfileTemplate_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    type: string;
}>;
export type AppDeploymentSettings_GetDockerfileTemplate_Res = ApiResponseBase<{ template: string }>;
