import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectCommandPipe_CreateOne_Req,
    ProjectCommandPipe_CreateOne_Res,
    ProjectCommandPipe_DeleteOne_Req,
    ProjectCommandPipe_DeleteOne_Res,
    ProjectCommandPipe_FindManyPaginated_Req,
    ProjectCommandPipe_FindManyPaginated_Res,
    ProjectCommandPipe_FindOneById_Req,
    ProjectCommandPipe_FindOneById_Res,
    ProjectCommandPipe_UpdateOne_Req,
    ProjectCommandPipe_UpdateOne_Res,
    ProjectCommandPipe_UpdateStatus_Req,
    ProjectCommandPipe_UpdateStatus_Res,
} from "./project-command-pipe.api.contracts";
import type { ProjectCommandPipeApiValidator } from "./project-command-pipe.api.validator";

export class ProjectCommandPipeApi extends BaseApi {
    public constructor(private readonly validator: ProjectCommandPipeApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectCommandPipe_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/command-pipes`, {
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
        request: ProjectCommandPipe_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/command-pipes/${id}`, {
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
        request: ProjectCommandPipe_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/command-pipes`, payload, {
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
        request: ProjectCommandPipe_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/command-pipes/${id}`, payload, {
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
        request: ProjectCommandPipe_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/command-pipes/${id}/status`, payload, {
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
        request: ProjectCommandPipe_DeleteOne_Req,
    ): Promise<Result<ProjectCommandPipe_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/command-pipes/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
