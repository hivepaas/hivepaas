import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    PUBLIC_LIST_ALL,
    PUBLIC_LIST_STATUS_ACTIVE,
    type ProjectsPublicApiValidator,
    type Public_Projects_FindManyPaginated_Req,
    type Public_Projects_FindManyPaginated_Res,
} from "@application/shared/api-public/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class ProjectsPublicApi extends BaseApi {
    public constructor(private readonly validator: ProjectsPublicApiValidator) {
        super();
    }

    /**
     * The active projects the user can see, by name: every one of them unless a
     * page is asked for.
     */
    async findManyPaginated(
        request: Public_Projects_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Projects_FindManyPaginated_Res, Error>> {
        const { search, pagination } = request.data;

        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination ?? PUBLIC_LIST_ALL)
            .sorting([{ id: "name", desc: false }])
            .search(search)
            .filterBy({ status: [PUBLIC_LIST_STATUS_ACTIVE] });

        return lastValueFrom(
            from(
                this.client.v1.get("/projects", {
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
