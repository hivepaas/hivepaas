import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppDeploymentSettings_FindOne_Req,
    type AppDeploymentSettings_FindOne_Res,
    type AppDeploymentSettings_UpdateOne_Req,
    type AppDeploymentSettings_UpdateOne_Res,
} from "./app-deployment-settings.api.contracts";
import { type AppDeploymentSettingsApiValidator } from "./app-deployment-settings.api.validator";

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

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/apps/${appID}/deployment-settings`,
                    { ...payload, updateVer },
                    {
                        signal,
                    },
                ),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
