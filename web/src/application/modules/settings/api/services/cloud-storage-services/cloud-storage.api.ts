import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    CloudStorage_CreateOne_Req,
    CloudStorage_CreateOne_Res,
    CloudStorage_DeleteOne_Req,
    CloudStorage_DeleteOne_Res,
    CloudStorage_FindManyPaginated_Req,
    CloudStorage_FindManyPaginated_Res,
    CloudStorage_FindOneById_Req,
    CloudStorage_FindOneById_Res,
    CloudStorage_TestConn_Req,
    CloudStorage_TestConn_Res,
    CloudStorage_UpdateMeta_Req,
    CloudStorage_UpdateMeta_Res,
    CloudStorage_UpdateOne_Req,
    CloudStorage_UpdateOne_Res,
} from "./cloud-storage.api.contracts";
import type { CloudStorageApiValidator } from "./cloud-storage.api.validator";

export class CloudStorageApi extends BaseApi {
    public constructor(private readonly validator: CloudStorageApiValidator) {
        super();
    }

    async findManyPaginated(
        request: CloudStorage_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(this.client.v1.get("/settings/cloud-storages", { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: CloudStorage_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/settings/cloud-storages/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: CloudStorage_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/cloud-storages", payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: CloudStorage_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/cloud-storages/${id}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: CloudStorage_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_UpdateMeta_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/settings/cloud-storages/${id}/status`, payload, { signal })).pipe(
                map(this.validator.updateMeta),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: CloudStorage_DeleteOne_Req): Promise<Result<CloudStorage_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/cloud-storages/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testConn(
        request: CloudStorage_TestConn_Req,
        signal?: AbortSignal,
    ): Promise<Result<CloudStorage_TestConn_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/settings/cloud-storages/test-conn", payload, { signal })).pipe(
                map(this.validator.testConn),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
