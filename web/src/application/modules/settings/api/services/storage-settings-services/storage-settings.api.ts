import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    StorageSettings_DeleteOne_Req,
    StorageSettings_DeleteOne_Res,
    StorageSettings_FindOne_Req,
    StorageSettings_FindOne_Res,
    StorageSettings_UpdateOne_Req,
    StorageSettings_UpdateOne_Res,
    StorageSettings_UpdateStatus_Req,
    StorageSettings_UpdateStatus_Res,
} from "./storage-settings.api.contracts";
import type { StorageSettingsApiValidator } from "./storage-settings.api.validator";

export class StorageSettingsApi extends BaseApi {
    public constructor(private readonly validator: StorageSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: StorageSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<StorageSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/settings/storage-settings", {
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
        request: StorageSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<StorageSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/storage-settings", payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: StorageSettings_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<StorageSettings_UpdateStatus_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/storage-settings/status", payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(_request: StorageSettings_DeleteOne_Req): Promise<Result<StorageSettings_DeleteOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.delete("/settings/storage-settings")).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
