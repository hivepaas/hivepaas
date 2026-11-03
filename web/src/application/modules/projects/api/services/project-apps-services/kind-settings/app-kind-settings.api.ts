import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppKindSettings_FindOne_Req,
    type AppKindSettings_FindOne_Res,
    type AppKindSettings_UpdateOne_Req,
    type AppKindSettings_UpdateOne_Res,
} from "./app-kind-settings.api.contracts";
import { type AppKindSettingsApiValidator } from "./app-kind-settings.api.validator";

export class AppKindSettingsApi extends BaseApi {
    constructor(private readonly validator: AppKindSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppKindSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppKindSettings_FindOne_Res, Error>> {
        const { projectID, env, appID, revealSecrets } = req.data;
        const query = this.queryBuilder.getInstance();
        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/kind-settings`, {
                    params: {
                        ...query.build(),
                        ...(revealSecrets === undefined ? {} : { revealSecrets }),
                    },
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
        req: AppKindSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppKindSettings_UpdateOne_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/kind-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
