import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectDomainSettings_DeleteOne_Req,
    ProjectDomainSettings_DeleteOne_Res,
    ProjectDomainSettings_FindOne_Req,
    ProjectDomainSettings_FindOne_Res,
    ProjectDomainSettings_UpdateOne_Req,
    ProjectDomainSettings_UpdateOne_Res,
    ProjectDomainSettings_UpdateStatus_Req,
    ProjectDomainSettings_UpdateStatus_Res,
} from "./project-domain-settings.api.contracts";
import type { ProjectDomainSettingsApiValidator } from "./project-domain-settings.api.validator";

export class ProjectDomainSettingsApi extends BaseApi {
    public constructor(private readonly validator: ProjectDomainSettingsApiValidator) {
        super();
    }

    async findOne(
        request: ProjectDomainSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectDomainSettings_FindOne_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/domain-settings`, {
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
        request: ProjectDomainSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectDomainSettings_UpdateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/domain-settings`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectDomainSettings_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectDomainSettings_UpdateStatus_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/domain-settings/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectDomainSettings_DeleteOne_Req,
    ): Promise<Result<ProjectDomainSettings_DeleteOne_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/domain-settings`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
