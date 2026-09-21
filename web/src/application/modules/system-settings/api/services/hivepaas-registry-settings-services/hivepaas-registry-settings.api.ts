import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSRegistrySettings_CheckPush_Req,
    HivePaaSRegistrySettings_CheckPush_Res,
    HivePaaSRegistrySettings_FindOne_Req,
    HivePaaSRegistrySettings_FindOne_Res,
    HivePaaSRegistrySettings_ProbeDomain_Req,
    HivePaaSRegistrySettings_ProbeDomain_Res,
    HivePaaSRegistrySettings_RotateCredential_Req,
    HivePaaSRegistrySettings_RotateCredential_Res,
    HivePaaSRegistrySettings_UpdateOne_Req,
    HivePaaSRegistrySettings_UpdateOne_Res,
} from "./hivepaas-registry-settings.api.contracts";
import type { HivePaaSRegistrySettingsApiValidator } from "./hivepaas-registry-settings.api.validator";

/** The check uploads 150 MB and waits for the answer, which no default timeout allows for. */
const PUSH_CHECK_TIMEOUT_MS = 10 * 60 * 1000;

export class HivePaaSRegistrySettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSRegistrySettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSRegistrySettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRegistrySettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/registry", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSRegistrySettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRegistrySettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        // Saving also provisions the registry, or reconciles the one that exists,
        // so this can take several seconds.
        return lastValueFrom(
            from(this.client.v1.put("/system/settings/registry", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async probeDomain(
        request: HivePaaSRegistrySettings_ProbeDomain_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRegistrySettings_ProbeDomain_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/registry/probe-domain", request.data, { signal })).pipe(
                map(this.validator.probeDomain),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async checkPush(
        request: HivePaaSRegistrySettings_CheckPush_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRegistrySettings_CheckPush_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/system/settings/registry/push-check", request.data, {
                    signal,
                    timeout: PUSH_CHECK_TIMEOUT_MS,
                }),
            ).pipe(
                map(this.validator.checkPush),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async rotateCredential(
        _request: HivePaaSRegistrySettings_RotateCredential_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRegistrySettings_RotateCredential_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/registry/rotate-credential", {}, { signal })).pipe(
                map(this.validator.rotateCredential),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
