import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    GetStarted_Dismiss_Req,
    GetStarted_Dismiss_Res,
    GetStarted_RequestDashboardCert_Req,
    GetStarted_RequestDashboardCert_Res,
} from "./get-started.api.contracts";
import type { GetStartedApiValidator } from "./get-started.api.validator";

export class GetStartedApi extends BaseApi {
    public constructor(private readonly validator: GetStartedApiValidator) {
        super();
    }

    async requestDashboardCert(
        _request: GetStarted_RequestDashboardCert_Req,
        signal?: AbortSignal,
    ): Promise<Result<GetStarted_RequestDashboardCert_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/get-started/dashboard-cert", {}, { signal })).pipe(
                map(this.validator.requestDashboardCert),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async dismiss(
        _request: GetStarted_Dismiss_Req,
        signal?: AbortSignal,
    ): Promise<Result<GetStarted_Dismiss_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/get-started/dismiss", {}, { signal })).pipe(
                map(this.validator.dismiss),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
