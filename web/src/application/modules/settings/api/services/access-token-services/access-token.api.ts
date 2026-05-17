import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AccessToken_CreateOne_Req,
    AccessToken_CreateOne_Res,
    AccessToken_DeleteOne_Req,
    AccessToken_DeleteOne_Res,
    AccessToken_FindManyPaginated_Req,
    AccessToken_FindManyPaginated_Res,
    AccessToken_FindOneById_Req,
    AccessToken_FindOneById_Res,
    AccessToken_TestConn_Req,
    AccessToken_TestConn_Res,
    AccessToken_UpdateMeta_Req,
    AccessToken_UpdateMeta_Res,
    AccessToken_UpdateOne_Req,
    AccessToken_UpdateOne_Res,
} from "./access-token.api.contracts";
import type { AccessTokenApiValidator } from "./access-token.api.validator";

export class AccessTokenApi extends BaseApi {
    public constructor(private readonly validator: AccessTokenApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AccessToken_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get("/settings/access-tokens", { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: AccessToken_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/settings/access-tokens/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: AccessToken_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/access-tokens", payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: AccessToken_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/access-tokens/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: AccessToken_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_UpdateMeta_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/access-tokens/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AccessToken_DeleteOne_Req): Promise<Result<AccessToken_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/access-tokens/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testConn(
        request: AccessToken_TestConn_Req,
        signal?: AbortSignal,
    ): Promise<Result<AccessToken_TestConn_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/access-tokens/test-conn", payload, { signal })).pipe(
                map(this.validator.testConn),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
