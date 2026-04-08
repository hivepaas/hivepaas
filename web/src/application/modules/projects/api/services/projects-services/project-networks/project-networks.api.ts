import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type ProjectNetworks_FindManyPaginated_Req,
    type ProjectNetworks_FindManyPaginated_Res,
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
}
