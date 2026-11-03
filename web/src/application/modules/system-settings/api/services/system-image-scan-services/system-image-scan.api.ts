import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemImageScan_Execute_Req,
    SystemImageScan_Execute_Res,
    SystemImageScan_FindApps_Req,
    SystemImageScan_FindApps_Res,
    SystemImageScan_FindComponent_Req,
    SystemImageScan_FindComponent_Res,
    SystemImageScan_FindComponents_Req,
    SystemImageScan_FindComponents_Res,
    SystemImageScan_FindOne_Req,
    SystemImageScan_FindOne_Res,
    SystemImageScan_UpdateOne_Req,
    SystemImageScan_UpdateOne_Res,
} from "./system-image-scan.api.contracts";
import type { SystemImageScanApiValidator } from "./system-image-scan.api.validator";

export class SystemImageScanApi extends BaseApi {
    public constructor(private readonly validator: SystemImageScanApiValidator) {
        super();
    }

    async findOne(
        _request: SystemImageScan_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/image-scan", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemImageScan_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/image-scan", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        _request: SystemImageScan_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/image-scan/exec", {}, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findApps(
        request: SystemImageScan_FindApps_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_FindApps_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/image-scan/apps", { params: request.data, signal })).pipe(
                map(this.validator.findApps),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findComponents(
        _request: SystemImageScan_FindComponents_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_FindComponents_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/image-scan/hivepaas", { signal })).pipe(
                map(this.validator.findComponents),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findComponent(
        request: SystemImageScan_FindComponent_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemImageScan_FindComponent_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get(`/system/settings/image-scan/hivepaas/${request.data.component}`, { signal })).pipe(
                map(this.validator.findComponent),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
