import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    PUBLIC_LIST_ALL,
    type Public_Users_FindMany_Req,
    type Public_Users_FindMany_Res,
    type UsersPublicApiValidator,
} from "@application/shared/api-public/services";
import { EUserStatus } from "@application/shared/enums";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class UsersPublicApi extends BaseApi {
    public constructor(private readonly validator: UsersPublicApiValidator) {
        super();
    }

    /**
     * The active users, by full name - every one of them, for a picker or a
     * filter to offer.
     */
    async findMany(
        request: Public_Users_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Users_FindMany_Res, Error>> {
        const { search, role } = request.data;

        const query = this.queryBuilder.getInstance();
        query
            .pagination(PUBLIC_LIST_ALL)
            .sorting([{ id: "full_name", desc: false }])
            .search(search)
            .filterBy({ status: [EUserStatus.Active], role: [role] });

        return lastValueFrom(
            from(
                this.client.v1.get("/users", {
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
