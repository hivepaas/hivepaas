import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    ClusterVolumes_CreateOne_Req,
    ClusterVolumes_CreateOne_Res,
    ClusterVolumes_DeleteOne_Req,
    ClusterVolumes_DeleteOne_Res,
    ClusterVolumes_FindManyPaginated_Req,
    ClusterVolumes_FindManyPaginated_Res,
    ClusterVolumes_FindOneById_Req,
    ClusterVolumes_FindOneById_Res,
    ClusterVolumes_List_Req,
    ClusterVolumes_List_Res,
    ClusterVolumes_SyncFromDocker_Req,
    ClusterVolumes_SyncFromDocker_Res,
    ClusterVolumes_UpdateOne_Req,
    ClusterVolumes_UpdateOne_Res,
    ClusterVolumes_UpdateStatus_Req,
    ClusterVolumes_UpdateStatus_Res,
} from "./volumes.api.contracts";
import type { ClusterVolumesApiValidator } from "./volumes.api.validator";

export class ClusterVolumesApi extends BaseApi {
    public constructor(private readonly validator: ClusterVolumesApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ClusterVolumes_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting, type } = request.data;
        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({ type: [type] });

        return lastValueFrom(
            from(
                this.client.v1.get("/cluster/volumes", {
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

    async list(
        request: ClusterVolumes_List_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_List_Res, Error>> {
        return this.findManyPaginated(request, signal);
    }

    async findOneById(
        request: ClusterVolumes_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_FindOneById_Res, Error>> {
        const { volumeID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/cluster/volumes/${volumeID}`, {
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
        request: ClusterVolumes_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_CreateOne_Res, Error>> {
        const { payload } = request.data;
        const json = {
            ...payload,
            name: JsonTransformer.string({ data: payload.name }),
            driver: JsonTransformer.string({ data: payload.driver }),
        };

        return lastValueFrom(
            from(this.client.v1.post("/cluster/volumes", json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ClusterVolumes_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_UpdateOne_Res, Error>> {
        const { volumeID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/cluster/volumes/${volumeID}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ClusterVolumes_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_UpdateStatus_Res, Error>> {
        const { volumeID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/cluster/volumes/${volumeID}/status`, payload, { signal })).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ClusterVolumes_DeleteOne_Req): Promise<Result<ClusterVolumes_DeleteOne_Res, Error>> {
        const { volumeID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/cluster/volumes/${volumeID}`)).pipe(
                map(() =>
                    Ok({
                        data: {
                            type: "success" as const,
                        },
                    }),
                ),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async syncFromDocker(
        _request: ClusterVolumes_SyncFromDocker_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_SyncFromDocker_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/cluster/volumes/sync", {}, { signal })).pipe(
                map(this.validator.syncFromDocker),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
