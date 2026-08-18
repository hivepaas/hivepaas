import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ImageBuildSettings_ClearRepoCache_Req,
    ImageBuildSettings_ClearRepoCache_Res,
    ImageBuildSettings_FindOne_Req,
    ImageBuildSettings_FindOne_Res,
    ImageBuildSettings_FindRepoCache_Req,
    ImageBuildSettings_FindRepoCache_Res,
    ImageBuildSettings_UpdateOne_Req,
    ImageBuildSettings_UpdateOne_Res,
} from "./image-build-settings.api.contracts";
import type { ImageBuildSettingsApiValidator } from "./image-build-settings.api.validator";

export class ImageBuildSettingsApi extends BaseApi {
    public constructor(private readonly validator: ImageBuildSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: ImageBuildSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImageBuildSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/settings/image-build-settings", {
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
        request: ImageBuildSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImageBuildSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put("/settings/image-build-settings", payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findRepoCache(
        _request: ImageBuildSettings_FindRepoCache_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImageBuildSettings_FindRepoCache_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/settings/image-build-settings/repo-cache", {
                    signal,
                }),
            ).pipe(
                map(this.validator.findRepoCache),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async clearRepoCache(
        _request: ImageBuildSettings_ClearRepoCache_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImageBuildSettings_ClearRepoCache_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/settings/image-build-settings/repo-cache/clear", {}, { signal })).pipe(
                map(this.validator.clearRepoCache),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
