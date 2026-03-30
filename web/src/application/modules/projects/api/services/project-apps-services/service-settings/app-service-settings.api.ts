import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppServiceSettings_FindOne_Req,
    type AppServiceSettings_FindOne_Res,
    type AppServiceSettings_UpdateOne_Req,
    type AppServiceSettings_UpdateOne_Res,
} from "./app-service-settings.api.contracts";
import { type AppServiceSettingsApiValidator } from "./app-service-settings.api.validator";

export class AppServiceSettingsApi extends BaseApi {
    constructor(private readonly validator: AppServiceSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppServiceSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppServiceSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/service-settings`, {
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
        req: AppServiceSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppServiceSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/service-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
