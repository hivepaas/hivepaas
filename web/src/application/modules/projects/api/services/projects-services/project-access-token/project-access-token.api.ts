import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectAccessToken_CreateOne_Req,
    ProjectAccessToken_CreateOne_Res,
    ProjectAccessToken_DeleteOne_Req,
    ProjectAccessToken_DeleteOne_Res,
    ProjectAccessToken_FindManyPaginated_Req,
    ProjectAccessToken_FindManyPaginated_Res,
    ProjectAccessToken_FindOneById_Req,
    ProjectAccessToken_FindOneById_Res,
    ProjectAccessToken_UpdateMeta_Req,
    ProjectAccessToken_UpdateMeta_Res,
    ProjectAccessToken_UpdateOne_Req,
    ProjectAccessToken_UpdateOne_Res,
} from "./project-access-token.api.contracts";
import type { ProjectAccessTokenApiValidator } from "./project-access-token.api.validator";

export class ProjectAccessTokenApi extends BaseApi {
    public constructor(private readonly validator: ProjectAccessTokenApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectAccessToken_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAccessToken_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/access-tokens`, { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectAccessToken_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAccessToken_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/access-tokens/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectAccessToken_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAccessToken_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/access-tokens`, payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectAccessToken_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAccessToken_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/access-tokens/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: ProjectAccessToken_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAccessToken_UpdateMeta_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/access-tokens/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectAccessToken_DeleteOne_Req,
    ): Promise<Result<ProjectAccessToken_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/access-tokens/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
