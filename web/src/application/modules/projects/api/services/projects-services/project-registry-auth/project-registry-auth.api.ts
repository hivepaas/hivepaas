import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { ProjectRegistryAuthApiValidator } from "./project-registry-auth.api.validator";
import type {
    ProjectRegistryAuth_CreateOne_Req,
    ProjectRegistryAuth_CreateOne_Res,
    ProjectRegistryAuth_DeleteOne_Req,
    ProjectRegistryAuth_DeleteOne_Res,
    ProjectRegistryAuth_FindManyPaginated_Req,
    ProjectRegistryAuth_FindManyPaginated_Res,
    ProjectRegistryAuth_FindOneById_Req,
    ProjectRegistryAuth_FindOneById_Res,
    ProjectRegistryAuth_UpdateMeta_Req,
    ProjectRegistryAuth_UpdateMeta_Res,
    ProjectRegistryAuth_UpdateOne_Req,
    ProjectRegistryAuth_UpdateOne_Res,
} from "./project-registry-auth.api.contracts";

export class ProjectRegistryAuthApi extends BaseApi {
    public constructor(private readonly validator: ProjectRegistryAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectRegistryAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/registry-auth`, {
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
        request: ProjectRegistryAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/registry-auth/${id}`, {
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
        request: ProjectRegistryAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/registry-auth`, payload, {
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
        request: ProjectRegistryAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/registry-auth/${id}`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: ProjectRegistryAuth_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_UpdateMeta_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/registry-auth/${id}/meta`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectRegistryAuth_DeleteOne_Req,
    ): Promise<Result<ProjectRegistryAuth_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/registry-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
