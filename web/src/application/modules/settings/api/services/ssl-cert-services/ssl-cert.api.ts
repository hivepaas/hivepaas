import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { SslCertApiValidator } from "./ssl-cert.api.validator";
import type {
    SslCert_CreateOne_Req,
    SslCert_CreateOne_Res,
    SslCert_DeleteOne_Req,
    SslCert_DeleteOne_Res,
    SslCert_FindManyPaginated_Req,
    SslCert_FindManyPaginated_Res,
    SslCert_FindOneById_Req,
    SslCert_FindOneById_Res,
    SslCert_UpdateOne_Req,
    SslCert_UpdateOne_Res,
    SslCert_UpdateStatus_Req,
    SslCert_UpdateStatus_Res,
} from "./ssl-cert.api.contracts";

export class SslCertApi extends BaseApi {
    public constructor(private readonly validator: SslCertApiValidator) {
        super();
    }

    async findManyPaginated(
        request: SslCert_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslCert_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting, domain } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(domain !== undefined && domain !== "" ? { domain } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/ssl-certs", {
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
        request: SslCert_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslCert_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/ssl-certs/${id}`, {
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
        request: SslCert_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslCert_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/ssl-certs", payload, {
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
        request: SslCert_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslCert_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/ssl-certs/${id}`, payload, {
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
        request: SslCert_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<SslCert_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/ssl-certs/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: SslCert_DeleteOne_Req): Promise<Result<SslCert_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/ssl-certs/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
