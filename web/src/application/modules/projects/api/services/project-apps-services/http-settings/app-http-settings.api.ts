import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppHttpSettings_FindOne_Req,
    type AppHttpSettings_FindOne_Res,
    type AppHttpSettings_UpdateOne_Req,
    type AppHttpSettings_UpdateOne_Res,
} from "./app-http-settings.api.contracts";
import { type AppHttpSettingsApiValidator } from "./app-http-settings.api.validator";

export class AppHttpSettingsApi extends BaseApi {
    constructor(private readonly validator: AppHttpSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppHttpSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHttpSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/http-settings`, {
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
        req: AppHttpSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHttpSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/http-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
