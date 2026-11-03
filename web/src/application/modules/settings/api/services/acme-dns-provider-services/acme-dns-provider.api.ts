import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AcmeDnsProvider_CreateOne_Req,
    AcmeDnsProvider_CreateOne_Res,
    AcmeDnsProvider_DeleteOne_Req,
    AcmeDnsProvider_DeleteOne_Res,
    AcmeDnsProvider_FindManyPaginated_Req,
    AcmeDnsProvider_FindManyPaginated_Res,
    AcmeDnsProvider_FindOneById_Req,
    AcmeDnsProvider_FindOneById_Res,
    AcmeDnsProvider_TestAccess_Req,
    AcmeDnsProvider_TestAccess_Res,
    AcmeDnsProvider_UpdateOne_Req,
    AcmeDnsProvider_UpdateOne_Res,
    AcmeDnsProvider_UpdateStatus_Req,
    AcmeDnsProvider_UpdateStatus_Res,
} from "./acme-dns-provider.api.contracts";
import type { AcmeDnsProviderApiValidator } from "./acme-dns-provider.api.validator";

export class AcmeDnsProviderApi extends BaseApi {
    public constructor(private readonly validator: AcmeDnsProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AcmeDnsProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting, kind } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(kind ? { kind } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/acme-dns-providers", {
                    params,
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
        request: AcmeDnsProvider_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/acme-dns-providers/${id}`, {
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
        request: AcmeDnsProvider_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/acme-dns-providers", payload, {
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
        request: AcmeDnsProvider_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/acme-dns-providers/${id}`, payload, {
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
        request: AcmeDnsProvider_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/acme-dns-providers/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AcmeDnsProvider_DeleteOne_Req): Promise<Result<AcmeDnsProvider_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/acme-dns-providers/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testAccess(
        request: AcmeDnsProvider_TestAccess_Req,
        signal?: AbortSignal,
    ): Promise<Result<AcmeDnsProvider_TestAccess_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/acme-dns-providers/test-access", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.testAccess),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
