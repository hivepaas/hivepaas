import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppDockerApiSettings_FindOne_Req,
    AppDockerApiSettings_FindOne_Res,
    AppDockerApiSettings_UpdateOne_Req,
    AppDockerApiSettings_UpdateOne_Res,
} from "./app-docker-api-settings.api.contracts";
import type { AppDockerApiSettingsApiValidator } from "./app-docker-api-settings.api.validator";

export class AppDockerApiSettingsApi extends BaseApi {
    constructor(private readonly validator: AppDockerApiSettingsApiValidator) {
        super();
    }

    async findOne(
        req: AppDockerApiSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDockerApiSettings_FindOne_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/docker-api-settings`, { signal }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        req: AppDockerApiSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDockerApiSettings_UpdateOne_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/docker-api-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
