import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppCloneSettings_Execute_Req,
    AppCloneSettings_Execute_Res,
    AppCloneSettings_FindOne_Req,
    AppCloneSettings_FindOne_Res,
    AppCloneSettings_UpdateOne_Req,
    AppCloneSettings_UpdateOne_Res,
} from "./app-clone-settings.api.contracts";
import type { AppCloneSettingsApiValidator } from "./app-clone-settings.api.validator";

export class AppCloneSettingsApi extends BaseApi {
    constructor(private readonly validator: AppCloneSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppCloneSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppCloneSettings_FindOne_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/clone-settings`, {
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
        req: AppCloneSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppCloneSettings_UpdateOne_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/clone-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        req: AppCloneSettings_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppCloneSettings_Execute_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/${env}/apps/${appID}/clone-execute`, {}, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
