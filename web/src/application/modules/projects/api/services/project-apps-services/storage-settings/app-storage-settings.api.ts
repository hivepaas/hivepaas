import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppStorageSettings_FindOne_Req,
    type AppStorageSettings_FindOne_Res,
    type AppStorageSettings_UpdateOne_Req,
    type AppStorageSettings_UpdateOne_Res,
} from "./app-storage-settings.api.contracts";
import { type AppStorageSettingsApiValidator } from "./app-storage-settings.api.validator";

export class AppStorageSettingsApi extends BaseApi {
    constructor(private readonly validator: AppStorageSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppStorageSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppStorageSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/storage-settings`, {
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
        req: AppStorageSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppStorageSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/storage-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
