import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type AppContainerSettings_FindOne_Req,
    type AppContainerSettings_FindOne_Res,
    type AppContainerSettings_UpdateOne_Req,
    type AppContainerSettings_UpdateOne_Res,
} from "./app-container-settings.api.contracts";
import { type AppContainerSettingsApiValidator } from "./app-container-settings.api.validator";

export class AppContainerSettingsApi extends BaseApi {
    constructor(private readonly validator: AppContainerSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppContainerSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/container-settings`, {
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
        req: AppContainerSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/container-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
