import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppRoutingSettings_FindOne_Req,
    type AppRoutingSettings_FindOne_Res,
    type AppRoutingSettings_UpdateOne_Req,
    type AppRoutingSettings_UpdateOne_Res,
} from "./app-routing-settings.api.contracts";
import { type AppRoutingSettingsApiValidator } from "./app-routing-settings.api.validator";

export class AppRoutingSettingsApi extends BaseApi {
    constructor(private readonly validator: AppRoutingSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppRoutingSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppRoutingSettings_FindOne_Res, Error>> {
        const { projectID, env, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/routing-settings`, {
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
        req: AppRoutingSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppRoutingSettings_UpdateOne_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/routing-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}

export { AppRoutingSettingsApi as AppHttpSettingsApi };
