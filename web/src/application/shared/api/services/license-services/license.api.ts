import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    License_FindOne_Req,
    License_FindOne_Res,
    License_InstallKey_Req,
    License_InstallKey_Res,
    License_Refresh_Req,
    License_Refresh_Res,
} from "./license.api.contracts";
import type { LicenseApiValidator } from "./license.api.validator";

/**
 * The installation's license. Shared rather than System's own: System › License,
 * the banners on every page, the plan-limit dialog and the OAuth page all read it.
 */
export class LicenseApi extends BaseApi {
    public constructor(private readonly validator: LicenseApiValidator) {
        super();
    }

    async findOne(_request: License_FindOne_Req, signal?: AbortSignal): Promise<Result<License_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/license", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * A key that does not verify, is for another installation, or is past its
     * grace is refused, and the stored key is kept. Whitespace in the key is
     * ignored by the server.
     */
    async installKey(
        request: License_InstallKey_Req,
        signal?: AbortSignal,
    ): Promise<Result<License_InstallKey_Res, Error>> {
        const { key } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/license", { key }, { signal })).pipe(
                map(this.validator.installKey),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Asks whether or not the daily check is on. A portal that cannot be reached
     * fails, and the failure is recorded as the last check's error.
     */
    async refresh(_request: License_Refresh_Req, signal?: AbortSignal): Promise<Result<License_Refresh_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/license/refresh", {}, { signal })).pipe(
                map(this.validator.refresh),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
