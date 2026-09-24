import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSUpdates_FindPlan_Req,
    HivePaaSUpdates_FindPlan_Res,
    HivePaaSUpdates_FindReleaseInfo_Req,
    HivePaaSUpdates_FindReleaseInfo_Res,
    HivePaaSUpdates_Update_Req,
    HivePaaSUpdates_Update_Res,
} from "./hivepaas-updates.api.contracts";
import type { HivePaaSUpdatesApiValidator } from "./hivepaas-updates.api.validator";

/** How long a single "is HivePaaS back" check may take before it counts as no. */
const PING_TIMEOUT_MS = 5_000;

export class HivePaaSUpdatesApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSUpdatesApiValidator) {
        super();
    }

    async findReleaseInfo(
        _request: HivePaaSUpdates_FindReleaseInfo_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSUpdates_FindReleaseInfo_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/release-info", { signal })).pipe(
                map(this.validator.findReleaseInfo),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /** What an update to that version would do, component by component. Nothing is changed. */
    async findPlan(
        request: HivePaaSUpdates_FindPlan_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSUpdates_FindPlan_Res, Error>> {
        const { targetVersion } = request.data;

        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/update-plan", { params: { targetVersion }, signal })).pipe(
                map(this.validator.findPlan),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async update(
        request: HivePaaSUpdates_Update_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSUpdates_Update_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/hivepaas/update-version", request.data, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Whether HivePaaS answers at all. It is what an update is watched with:
     * while the update runs, the app is stopped and nothing else can be asked.
     */
    async ping(): Promise<boolean> {
        try {
            await this.client.v1.get("/ping", { timeout: PING_TIMEOUT_MS });
            return true;
        } catch {
            return false;
        }
    }
}
