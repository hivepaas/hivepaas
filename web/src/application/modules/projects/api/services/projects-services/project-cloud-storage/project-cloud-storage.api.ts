import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectCloudStorage_CreateOne_Req,
    ProjectCloudStorage_CreateOne_Res,
    ProjectCloudStorage_DeleteOne_Req,
    ProjectCloudStorage_DeleteOne_Res,
    ProjectCloudStorage_FindManyPaginated_Req,
    ProjectCloudStorage_FindManyPaginated_Res,
    ProjectCloudStorage_FindOneById_Req,
    ProjectCloudStorage_FindOneById_Res,
    ProjectCloudStorage_UpdateMeta_Req,
    ProjectCloudStorage_UpdateMeta_Res,
    ProjectCloudStorage_UpdateOne_Req,
    ProjectCloudStorage_UpdateOne_Res,
} from "./project-cloud-storage.api.contracts";
import type { ProjectCloudStorageApiValidator } from "./project-cloud-storage.api.validator";

export class ProjectCloudStorageApi extends BaseApi {
    public constructor(private readonly validator: ProjectCloudStorageApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectCloudStorage_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/cloud-storages`, { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectCloudStorage_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/cloud-storages/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectCloudStorage_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/cloud-storages`, payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectCloudStorage_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/cloud-storages/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: ProjectCloudStorage_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_UpdateMeta_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/cloud-storages/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectCloudStorage_DeleteOne_Req,
    ): Promise<Result<ProjectCloudStorage_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/cloud-storages/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
