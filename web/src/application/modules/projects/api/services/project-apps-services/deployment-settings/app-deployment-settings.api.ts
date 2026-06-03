import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppDeploymentSettings_FindOne_Req,
    type AppDeploymentSettings_FindOne_Res,
    type AppDeploymentSettings_UpdateOne_Req,
    type AppDeploymentSettings_UpdateOne_Res,
} from "./app-deployment-settings.api.contracts";
import { type AppDeploymentSettingsApiValidator } from "./app-deployment-settings.api.validator";

function splitImageTags(imageTags: string | undefined): string[] {
    if (!imageTags) {
        return [];
    }

    return imageTags
        .split(",")
        .map(tag => tag.trim())
        .filter(Boolean);
}

export class AppDeploymentSettingsApi extends BaseApi {
    constructor(private readonly validator: AppDeploymentSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppDeploymentSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDeploymentSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/deployment-settings`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        req: AppDeploymentSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDeploymentSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, updateVer, payload } = req.data;
        const json =
            payload.activeMethod === EAppDeploymentMethod.Repo
                ? {
                      ...payload,
                      updateVer,
                      repoSource: {
                          buildTool: payload.repoSource.buildTool,
                          repoType: payload.repoSource.repoType,
                          repoURL: payload.repoSource.repoUrl,
                          repoRef: payload.repoSource.repoRef,
                          commitHash: payload.repoSource.commitHash,
                          repoOptions: payload.repoSource.repoOptions,
                          credentials: payload.repoSource.credentials,
                          dockerfilePath: payload.repoSource.dockerfilePath ?? "",
                          imageName: payload.repoSource.imageName,
                          imageTags: splitImageTags(payload.repoSource.imageTags),
                          pushToRegistry: payload.repoSource.pushToRegistry,
                      },
                  }
                : { ...payload, updateVer };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/deployment-settings`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
