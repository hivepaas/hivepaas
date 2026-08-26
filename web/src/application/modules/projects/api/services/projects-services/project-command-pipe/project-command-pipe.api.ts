import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectCommandPipe_CreateFromTemplate_Req,
    ProjectCommandPipe_CreateFromTemplate_Res,
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

function getProjectCommandPipeBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/command-pipes`;
    }

    return `/projects/${projectID}/command-pipes`;
}

export class ProjectCommandPipeApi extends BaseApi {
    public constructor(private readonly validator: ProjectCommandPipeApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectCommandPipe_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectCommandPipeBasePath(projectID, env), {
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
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectCommandPipeBasePath(projectID, env)}/${id}`, {
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
        const { projectID, env, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectCommandPipeBasePath(projectID, env), json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createFromTemplate(
        request: ProjectCommandPipe_CreateFromTemplate_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_CreateFromTemplate_Res, Error>> {
        const { projectID, env, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`${getProjectCommandPipeBasePath(projectID, env)}/from-template`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createFromTemplate),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectCommandPipe_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCommandPipe_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectCommandPipeBasePath(projectID, env)}/${id}`, json, {
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
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectCommandPipeBasePath(projectID, env)}/${id}/status`, json, {
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
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectCommandPipeBasePath(projectID, env)}/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
