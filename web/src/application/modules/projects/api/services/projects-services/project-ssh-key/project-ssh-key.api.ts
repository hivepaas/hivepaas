import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectSSHKey_CreateOne_Req,
    ProjectSSHKey_CreateOne_Res,
    ProjectSSHKey_DeleteOne_Req,
    ProjectSSHKey_DeleteOne_Res,
    ProjectSSHKey_FindManyPaginated_Req,
    ProjectSSHKey_FindManyPaginated_Res,
    ProjectSSHKey_FindOneById_Req,
    ProjectSSHKey_FindOneById_Res,
    ProjectSSHKey_UpdateMeta_Req,
    ProjectSSHKey_UpdateMeta_Res,
    ProjectSSHKey_UpdateOne_Req,
    ProjectSSHKey_UpdateOne_Res,
} from "./project-ssh-key.api.contracts";
import type { ProjectSSHKeyApiValidator } from "./project-ssh-key.api.validator";

export class ProjectSSHKeyApi extends BaseApi {
    public constructor(private readonly validator: ProjectSSHKeyApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectSSHKey_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSSHKey_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/ssh-keys`, { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectSSHKey_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSSHKey_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/ssh-keys/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectSSHKey_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSSHKey_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/ssh-keys`, payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectSSHKey_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSSHKey_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/ssh-keys/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: ProjectSSHKey_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSSHKey_UpdateMeta_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/ssh-keys/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectSSHKey_DeleteOne_Req): Promise<Result<ProjectSSHKey_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/ssh-keys/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
