import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppPlacementSettings_FindOne_Req,
    AppPlacementSettings_FindOne_Res,
    AppPlacementSettings_UpdateOne_Req,
    AppPlacementSettings_UpdateOne_Res,
} from "./app-placement-settings.api.contracts";
import type { AppPlacementSettingsApiValidator } from "./app-placement-settings.api.validator";

export class AppPlacementSettingsApi extends BaseApi {
    public constructor(private readonly validator: AppPlacementSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: AppPlacementSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppPlacementSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/settings/app-placement-settings", {
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
        request: AppPlacementSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppPlacementSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/app-placement-settings", payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
