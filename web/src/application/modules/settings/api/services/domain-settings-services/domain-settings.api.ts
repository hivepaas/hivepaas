import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { DomainSettingsApiValidator } from "./domain-settings.api.validator";
import type {
    DomainSettings_DeleteOne_Req,
    DomainSettings_DeleteOne_Res,
    DomainSettings_FindOne_Req,
    DomainSettings_FindOne_Res,
    DomainSettings_UpdateOne_Req,
    DomainSettings_UpdateOne_Res,
    DomainSettings_UpdateStatus_Req,
    DomainSettings_UpdateStatus_Res,
} from "./domain-settings.api.contracts";

export class DomainSettingsApi extends BaseApi {
    public constructor(private readonly validator: DomainSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: DomainSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<DomainSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/settings/domain-settings", {
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
        request: DomainSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<DomainSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/domain-settings", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: DomainSettings_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<DomainSettings_UpdateStatus_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/domain-settings/status", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(_request: DomainSettings_DeleteOne_Req): Promise<Result<DomainSettings_DeleteOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.delete("/settings/domain-settings")).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
