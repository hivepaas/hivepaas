import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppFeatureSettings_FindOne_Req,
    AppFeatureSettings_FindOne_Res,
    AppFeatureSettings_UpdateOne_Req,
    AppFeatureSettings_UpdateOne_Res,
} from "./app-feature-settings.api.contracts";
import type { AppFeatureSettingsApiValidator } from "./app-feature-settings.api.validator";

export class AppFeatureSettingsApi extends BaseApi {
    constructor(private readonly validator: AppFeatureSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppFeatureSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppFeatureSettings_FindOne_Res, Error>> {
        const { projectID, appID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/feature-settings`, {
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
        req: AppFeatureSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppFeatureSettings_UpdateOne_Res, Error>> {
        const { projectID, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/feature-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
