import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { ClusterVolumes_List_Req, ClusterVolumes_List_Res } from "./volumes.api.contracts";
import type { ClusterVolumesApiValidator } from "./volumes.api.validator";

export class ClusterVolumesApi extends BaseApi {
    public constructor(private readonly validator: ClusterVolumesApiValidator) {
        super();
    }

    async list(
        request: ClusterVolumes_List_Req,
        signal?: AbortSignal,
    ): Promise<Result<ClusterVolumes_List_Res, Error>> {
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
                map(this.validator.list),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
