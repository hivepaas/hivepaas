import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import {
    type ProjectNetworks_CreateOne_Req,
    type ProjectNetworks_CreateOne_Res,
    type ProjectNetworks_DeleteOne_Req,
    type ProjectNetworks_DeleteOne_Res,
    type ProjectNetworks_FindManyPaginated_Req,
    type ProjectNetworks_FindManyPaginated_Res,
    type ProjectNetworks_FindOneById_Req,
    type ProjectNetworks_FindOneById_Res,
} from "./project-networks.api.contracts";
import { type ProjectNetworksApiValidator } from "./project-networks.api.validator";

export class ProjectNetworksApi extends BaseApi {
    public constructor(private readonly validator: ProjectNetworksApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectNetworks_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNetworks_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/docker-networks`, {
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
        request: ProjectNetworks_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNetworks_FindOneById_Res, Error>> {
        const { projectID, networkID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/docker-networks/${networkID}`, {
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
        request: ProjectNetworks_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNetworks_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;
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
        };

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/docker-networks`, json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectNetworks_DeleteOne_Req): Promise<Result<ProjectNetworks_DeleteOne_Res, Error>> {
        const { projectID, networkID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/docker-networks/${networkID}`)).pipe(
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
