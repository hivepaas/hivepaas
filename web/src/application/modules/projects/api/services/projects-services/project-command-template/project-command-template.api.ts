import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectCommandTemplate_CreateOne_Req,
    ProjectCommandTemplate_CreateOne_Res,
    ProjectCommandTemplate_DeleteOne_Req,
    ProjectCommandTemplate_DeleteOne_Res,
    ProjectCommandTemplate_FindManyPaginated_Req,
    ProjectCommandTemplate_FindManyPaginated_Res,
    ProjectCommandTemplate_FindOneById_Req,
    ProjectCommandTemplate_FindOneById_Res,
    ProjectCommandTemplate_UpdateOne_Req,
    ProjectCommandTemplate_UpdateOne_Res,
    ProjectCommandTemplate_UpdateStatus_Req,
    ProjectCommandTemplate_UpdateStatus_Res,
} from "./project-command-template.api.contracts";
import type { ProjectCommandTemplateApiValidator } from "./project-command-template.api.validator";

export class ProjectCommandTemplateApi extends BaseApi {
    public constructor(private readonly validator: ProjectCommandTemplateApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectCommandTemplate_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandTemplate_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/command-templates`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectCommandTemplate_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandTemplate_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/command-templates/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectCommandTemplate_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandTemplate_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/command-templates`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectCommandTemplate_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandTemplate_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/command-templates/${id}`, payload, {
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
        request: ProjectCommandTemplate_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandTemplate_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/command-templates/${id}/status`, payload, {
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
        request: ProjectCommandTemplate_DeleteOne_Req,
    ): Promise<Result<ProjectCommandTemplate_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/command-templates/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
