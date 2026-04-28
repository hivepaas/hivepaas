import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectStorageSettings_DeleteOne_Req,
    ProjectStorageSettings_DeleteOne_Res,
    ProjectStorageSettings_FindOne_Req,
    ProjectStorageSettings_FindOne_Res,
    ProjectStorageSettings_UpdateOne_Req,
    ProjectStorageSettings_UpdateOne_Res,
    ProjectStorageSettings_UpdateStatus_Req,
    ProjectStorageSettings_UpdateStatus_Res,
} from "./project-storage-settings.api.contracts";
import type { ProjectStorageSettingsApiValidator } from "./project-storage-settings.api.validator";

export class ProjectStorageSettingsApi extends BaseApi {
    constructor(private readonly validator: ProjectStorageSettingsApiValidator) {
        super();
    }

    async findOne(
        req: ProjectStorageSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectStorageSettings_FindOne_Res, Error>> {
        const { projectID } = req.data;
        const query = this.queryBuilder.getInstance();

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/storage-settings`, {
                    params: query.build(),
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
        request: ProjectStorageSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectStorageSettings_UpdateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/storage-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectStorageSettings_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectStorageSettings_UpdateStatus_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/storage-settings/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectStorageSettings_DeleteOne_Req,
    ): Promise<Result<ProjectStorageSettings_DeleteOne_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/storage-settings`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
