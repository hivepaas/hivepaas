import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppNetworkSettings_FindOne_Req,
    type AppNetworkSettings_FindOne_Res,
    type AppNetworkSettings_UpdateOne_Req,
    type AppNetworkSettings_UpdateOne_Res,
} from "./app-network-settings.api.contracts";
import { type AppNetworkSettingsApiValidator } from "./app-network-settings.api.validator";

export class AppNetworkSettingsApi extends BaseApi {
    constructor(private readonly validator: AppNetworkSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppNetworkSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppNetworkSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/network-settings`, {
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
        req: AppNetworkSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppNetworkSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/network-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
