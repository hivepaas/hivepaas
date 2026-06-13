import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SslProvider_CreateOne_Req,
    SslProvider_CreateOne_Res,
    SslProvider_DeleteOne_Req,
    SslProvider_DeleteOne_Res,
    SslProvider_FindManyPaginated_Req,
    SslProvider_FindManyPaginated_Res,
    SslProvider_FindOneById_Req,
    SslProvider_FindOneById_Res,
    SslProvider_UpdateOne_Req,
    SslProvider_UpdateOne_Res,
    SslProvider_UpdateStatus_Req,
    SslProvider_UpdateStatus_Res,
} from "./ssl-provider.api.contracts";
import type { SslProviderApiValidator } from "./ssl-provider.api.validator";

export class SslProviderApi extends BaseApi {
    public constructor(private readonly validator: SslProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: SslProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslProvider_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/ssl-providers", {
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
        request: SslProvider_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslProvider_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/ssl-providers/${id}`, {
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
        request: SslProvider_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslProvider_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/ssl-providers", payload, {
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
        request: SslProvider_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslProvider_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/ssl-providers/${id}`, payload, {
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
        request: SslProvider_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslProvider_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/ssl-providers/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: SslProvider_DeleteOne_Req): Promise<Result<SslProvider_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/ssl-providers/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
