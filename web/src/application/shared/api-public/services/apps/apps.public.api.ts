import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type AppsPublicApiValidator,
    PUBLIC_LIST_ALL,
    PUBLIC_LIST_STATUS_ACTIVE,
    type Public_Apps_FindMany_Req,
    type Public_Apps_FindMany_Res,
} from "@application/shared/api-public/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppsPublicApi extends BaseApi {
    public constructor(private readonly validator: AppsPublicApiValidator) {
        super();
    }

    /**
     * The active apps of a project the user can see, by name: every one of them
     * unless a page is asked for.
     */
    async findMany(
        request: Public_Apps_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Apps_FindMany_Res, Error>> {
        const { projectID, search, pagination } = request.data;

        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination ?? PUBLIC_LIST_ALL)
            .sorting([{ id: "name", desc: false }])
            .search(search)
            .filterBy({ status: [PUBLIC_LIST_STATUS_ACTIVE] });

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findMany),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
