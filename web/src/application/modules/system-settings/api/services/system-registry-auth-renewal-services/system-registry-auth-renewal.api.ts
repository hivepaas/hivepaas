import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemRegistryAuthRenewal_Execute_Req,
    SystemRegistryAuthRenewal_Execute_Res,
    SystemRegistryAuthRenewal_FindOne_Req,
    SystemRegistryAuthRenewal_FindOne_Res,
    SystemRegistryAuthRenewal_UpdateOne_Req,
    SystemRegistryAuthRenewal_UpdateOne_Res,
} from "./system-registry-auth-renewal.api.contracts";
import type { SystemRegistryAuthRenewalApiValidator } from "./system-registry-auth-renewal.api.validator";

export class SystemRegistryAuthRenewalApi extends BaseApi {
    public constructor(private readonly validator: SystemRegistryAuthRenewalApiValidator) {
        super();
    }

    async findOne(
        _request: SystemRegistryAuthRenewal_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemRegistryAuthRenewal_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/registry-auth-renewal", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemRegistryAuthRenewal_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemRegistryAuthRenewal_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/registry-auth-renewal", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        request: SystemRegistryAuthRenewal_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemRegistryAuthRenewal_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/registry-auth-renewal/exec", request.data, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
