import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppStorageSettings_FindOne_Req,
    type AppStorageSettings_FindOne_Res,
    type AppStorageSettings_Preflight_Req,
    type AppStorageSettings_Preflight_Res,
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
        const { projectID, env, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/storage-settings?getMounts=true`, {
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
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/storage-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Which of the mounts about to be saved reach a directory that already holds
     * something. It writes nothing.
     */
    async preflight(
        req: AppStorageSettings_Preflight_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppStorageSettings_Preflight_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/${env}/apps/${appID}/storage-settings/preflight`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.preflight),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
