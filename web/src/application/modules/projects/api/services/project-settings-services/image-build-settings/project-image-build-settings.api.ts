import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectImageBuildSettings_ClearRepoCache_Req,
    ProjectImageBuildSettings_ClearRepoCache_Res,
    ProjectImageBuildSettings_FindOne_Req,
    ProjectImageBuildSettings_FindOne_Res,
    ProjectImageBuildSettings_FindRepoCache_Req,
    ProjectImageBuildSettings_FindRepoCache_Res,
    ProjectImageBuildSettings_UpdateOne_Req,
    ProjectImageBuildSettings_UpdateOne_Res,
} from "./project-image-build-settings.api.contracts";
import type { ProjectImageBuildSettingsApiValidator } from "./project-image-build-settings.api.validator";

export class ProjectImageBuildSettingsApi extends BaseApi {
    constructor(private readonly validator: ProjectImageBuildSettingsApiValidator) {
        super();
    }

    async findOne(
        req: ProjectImageBuildSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImageBuildSettings_FindOne_Res, Error>> {
        const { projectID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/image-build-settings`, {
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
        req: ProjectImageBuildSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImageBuildSettings_UpdateOne_Res, Error>> {
        const { projectID, payload } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/image-build-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findRepoCache(
        req: ProjectImageBuildSettings_FindRepoCache_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImageBuildSettings_FindRepoCache_Res, Error>> {
        const { projectID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/image-build-settings/repo-cache`, {
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
        req: ProjectImageBuildSettings_ClearRepoCache_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImageBuildSettings_ClearRepoCache_Res, Error>> {
        const { projectID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/image-build-settings/repo-cache/clear`, {}, { signal }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
