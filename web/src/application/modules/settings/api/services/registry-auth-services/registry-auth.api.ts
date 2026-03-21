import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { RegistryAuthApiValidator } from "./registry-auth.api.validator";
import type {
    RegistryAuth_CreateOne_Req,
    RegistryAuth_CreateOne_Res,
    RegistryAuth_DeleteOne_Req,
    RegistryAuth_DeleteOne_Res,
    RegistryAuth_FindManyPaginated_Req,
    RegistryAuth_FindManyPaginated_Res,
    RegistryAuth_FindOneById_Req,
    RegistryAuth_FindOneById_Res,
    RegistryAuth_TestConn_Req,
    RegistryAuth_TestConn_Res,
    RegistryAuth_UpdateMeta_Req,
    RegistryAuth_UpdateMeta_Res,
    RegistryAuth_UpdateOne_Req,
    RegistryAuth_UpdateOne_Res,
} from "./registry-auth.api.contracts";

export class RegistryAuthApi extends BaseApi {
    public constructor(private readonly validator: RegistryAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: RegistryAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/registry-auth", {
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
        request: RegistryAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/registry-auth/${id}`, {
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
        request: RegistryAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/registry-auth", payload, {
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
        request: RegistryAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/registry-auth/${id}`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: RegistryAuth_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_UpdateMeta_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/registry-auth/${id}/meta`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: RegistryAuth_DeleteOne_Req): Promise<Result<RegistryAuth_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/registry-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testConn(
        request: RegistryAuth_TestConn_Req,
        signal?: AbortSignal,
    ): Promise<Result<RegistryAuth_TestConn_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/registry-auth/test-conn", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.testConn),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
