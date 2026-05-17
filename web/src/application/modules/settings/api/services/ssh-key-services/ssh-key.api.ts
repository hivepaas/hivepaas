import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SSHKey_CreateOne_Req,
    SSHKey_CreateOne_Res,
    SSHKey_DeleteOne_Req,
    SSHKey_DeleteOne_Res,
    SSHKey_FindManyPaginated_Req,
    SSHKey_FindManyPaginated_Res,
    SSHKey_FindOneById_Req,
    SSHKey_FindOneById_Res,
    SSHKey_UpdateMeta_Req,
    SSHKey_UpdateMeta_Res,
    SSHKey_UpdateOne_Req,
    SSHKey_UpdateOne_Res,
} from "./ssh-key.api.contracts";
import type { SSHKeyApiValidator } from "./ssh-key.api.validator";

export class SSHKeyApi extends BaseApi {
    public constructor(private readonly validator: SSHKeyApiValidator) {
        super();
    }

    async findManyPaginated(
        request: SSHKey_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SSHKey_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get("/settings/ssh-keys", { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: SSHKey_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<SSHKey_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/settings/ssh-keys/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(request: SSHKey_CreateOne_Req, signal?: AbortSignal): Promise<Result<SSHKey_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/ssh-keys", payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(request: SSHKey_UpdateOne_Req, signal?: AbortSignal): Promise<Result<SSHKey_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/ssh-keys/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: SSHKey_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<SSHKey_UpdateMeta_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/ssh-keys/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: SSHKey_DeleteOne_Req): Promise<Result<SSHKey_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/ssh-keys/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
