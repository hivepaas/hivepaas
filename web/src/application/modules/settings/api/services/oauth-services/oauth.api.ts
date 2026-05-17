import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    OAuth_CreateOne_Req,
    OAuth_CreateOne_Res,
    OAuth_DeleteOne_Req,
    OAuth_DeleteOne_Res,
    OAuth_FindManyPaginated_Req,
    OAuth_FindManyPaginated_Res,
    OAuth_FindOneById_Req,
    OAuth_FindOneById_Res,
    OAuth_UpdateMeta_Req,
    OAuth_UpdateMeta_Res,
    OAuth_UpdateOne_Req,
    OAuth_UpdateOne_Res,
} from "./oauth.api.contracts";
import type { OAuthApiValidator } from "./oauth.api.validator";

export class OAuthApi extends BaseApi {
    public constructor(private readonly validator: OAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: OAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<OAuth_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get("/settings/oauth", { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: OAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<OAuth_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/settings/oauth/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(request: OAuth_CreateOne_Req, signal?: AbortSignal): Promise<Result<OAuth_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/oauth", payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(request: OAuth_UpdateOne_Req, signal?: AbortSignal): Promise<Result<OAuth_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/oauth/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: OAuth_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<OAuth_UpdateMeta_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/oauth/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: OAuth_DeleteOne_Req): Promise<Result<OAuth_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/oauth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
