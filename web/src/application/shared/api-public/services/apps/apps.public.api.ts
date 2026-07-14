import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type AppsPublicApiValidator,
    type Public_Apps_FindManyBase_Req,
    type Public_Apps_FindManyBase_Res,
} from "@application/shared/api-public/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppsPublicApi extends BaseApi {
    public constructor(private readonly validator: AppsPublicApiValidator) {
        super();
    }

    /**
     * Find many public apps base
     */
    async findManyBase(
        request: Public_Apps_FindManyBase_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Apps_FindManyBase_Res, Error>> {
        const { projectID, search, pagination } = request.data;

        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting([{ id: "name", desc: false }])
            .search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/base`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyBase),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
