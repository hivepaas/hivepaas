import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppResourceSettings_FindOne_Req,
    type AppResourceSettings_FindOne_Res,
    type AppResourceSettings_UpdateOne_Req,
    type AppResourceSettings_UpdateOne_Res,
} from "./app-resource-settings.api.contracts";
import { type AppResourceSettingsApiValidator } from "./app-resource-settings.api.validator";

export class AppResourceSettingsApi extends BaseApi {
    constructor(private readonly validator: AppResourceSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppResourceSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppResourceSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/resource-settings`, {
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
        req: AppResourceSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppResourceSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/resource-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
