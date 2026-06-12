import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemSslRenewal_Execute_Req,
    SystemSslRenewal_Execute_Res,
    SystemSslRenewal_FindOne_Req,
    SystemSslRenewal_FindOne_Res,
    SystemSslRenewal_UpdateOne_Req,
    SystemSslRenewal_UpdateOne_Res,
} from "./system-ssl-renewal.api.contracts";
import type { SystemSslRenewalApiValidator } from "./system-ssl-renewal.api.validator";

export class SystemSslRenewalApi extends BaseApi {
    public constructor(private readonly validator: SystemSslRenewalApiValidator) {
        super();
    }

    async findOne(
        _request: SystemSslRenewal_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemSslRenewal_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/ssl-renewal", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemSslRenewal_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemSslRenewal_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/ssl-renewal", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        request: SystemSslRenewal_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemSslRenewal_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/ssl-renewal/exec", request.data, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
