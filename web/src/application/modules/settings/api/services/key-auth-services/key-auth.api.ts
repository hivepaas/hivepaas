import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    KeyAuth_CreateOne_Req,
    KeyAuth_CreateOne_Res,
    KeyAuth_DeleteOne_Req,
    KeyAuth_DeleteOne_Res,
    KeyAuth_FindManyPaginated_Req,
    KeyAuth_FindManyPaginated_Res,
    KeyAuth_FindOneById_Req,
    KeyAuth_FindOneById_Res,
    KeyAuth_UpdateOne_Req,
    KeyAuth_UpdateOne_Res,
    KeyAuth_UpdateStatus_Req,
    KeyAuth_UpdateStatus_Res,
} from "./key-auth.api.contracts";
import type { KeyAuthApiValidator } from "./key-auth.api.validator";

export class KeyAuthApi extends BaseApi {
    public constructor(private readonly validator: KeyAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: KeyAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<KeyAuth_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/key-auth", {
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
        request: KeyAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<KeyAuth_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/key-auth/${id}`, {
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
        request: KeyAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<KeyAuth_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/key-auth", payload, {
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
        request: KeyAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<KeyAuth_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/key-auth/${id}`, payload, {
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
        request: KeyAuth_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<KeyAuth_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/key-auth/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: KeyAuth_DeleteOne_Req): Promise<Result<KeyAuth_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/key-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
