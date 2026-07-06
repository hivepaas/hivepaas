import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ClusterNetworksApiValidator,
    ClusterNetworks_CreateOne_Req,
    ClusterNetworks_CreateOne_Res,
    ClusterNetworks_DeleteOne_Req,
    ClusterNetworks_DeleteOne_Res,
    ClusterNetworks_FindManyPaginated_Req,
    ClusterNetworks_FindManyPaginated_Res,
    ClusterNetworks_FindOneById_Req,
    ClusterNetworks_FindOneById_Res,
    ClusterNetworks_UpdateOne_Req,
    ClusterNetworks_UpdateOne_Res,
    ClusterNetworks_UpdateStatus_Req,
    ClusterNetworks_UpdateStatus_Res,
} from "~/cluster/api/services/network-services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ClusterNetworksApi extends BaseApi {
    public constructor(private readonly validator: ClusterNetworksApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ClusterNetworks_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterNetworks_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/cluster/networks", {
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
        request: ClusterNetworks_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterNetworks_FindOneById_Res, Error>> {
        const { networkID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/cluster/networks/${networkID}`, {
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
        request: ClusterNetworks_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterNetworks_CreateOne_Res, Error>> {
        const { payload } = request.data;
        const json = {
            name: JsonTransformer.string({ data: payload.name }),
            driver: payload.driver,
            enableIPv4: payload.enableIPv4,
            enableIPv6: payload.enableIPv6,
            internal: payload.internal,
            attachable: payload.attachable,
            ingress: payload.ingress,
            labels: payload.labels,
            options: payload.options,
            availableInProjects: payload.availableInProjects,
            default: payload.default ?? false,
        };

        return lastValueFrom(
            from(this.client.v1.post("/cluster/networks", json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ClusterNetworks_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterNetworks_UpdateOne_Res, Error>> {
        const { networkID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/cluster/networks/${networkID}`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ClusterNetworks_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterNetworks_UpdateStatus_Res, Error>> {
        const { networkID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/cluster/networks/${networkID}/status`, payload, { signal })).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ClusterNetworks_DeleteOne_Req): Promise<Result<ClusterNetworks_DeleteOne_Res, Error>> {
        const { networkID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/cluster/networks/${networkID}`)).pipe(
                map(() =>
                    Ok({
                        data: {
                            networkID,
                        },
                    }),
                ),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
