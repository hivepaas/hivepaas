import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { BasicAuthApiValidator } from "./basic-auth.api.validator";
import type {
    BasicAuth_CreateOne_Req,
    BasicAuth_CreateOne_Res,
    BasicAuth_DeleteOne_Req,
    BasicAuth_DeleteOne_Res,
    BasicAuth_FindManyPaginated_Req,
    BasicAuth_FindManyPaginated_Res,
    BasicAuth_FindOneById_Req,
    BasicAuth_FindOneById_Res,
    BasicAuth_UpdateOne_Req,
    BasicAuth_UpdateOne_Res,
    BasicAuth_UpdateStatus_Req,
    BasicAuth_UpdateStatus_Res,
} from "./basic-auth.api.contracts";

export class BasicAuthApi extends BaseApi {
    public constructor(private readonly validator: BasicAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: BasicAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<BasicAuth_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/basic-auth", {
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
        request: BasicAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<BasicAuth_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/basic-auth/${id}`, {
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
        request: BasicAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<BasicAuth_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/basic-auth", payload, {
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
        request: BasicAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<BasicAuth_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/basic-auth/${id}`, payload, {
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
        request: BasicAuth_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<BasicAuth_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/basic-auth/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: BasicAuth_DeleteOne_Req): Promise<Result<BasicAuth_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/basic-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
