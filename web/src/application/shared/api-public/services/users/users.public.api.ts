import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type Public_Users_FindManyBase_Req,
    type Public_Users_FindManyBase_Res,
    type UsersPublicApiValidator,
} from "@application/shared/api-public/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class UsersPublicApi extends BaseApi {
    public constructor(private readonly validator: UsersPublicApiValidator) {
        super();
    }

    /**
     * Find many public users base
     */
    async findManyBase(
        request: Public_Users_FindManyBase_Req,
        signal?: AbortSignal,
    ): Promise<Result<Public_Users_FindManyBase_Res, Error>> {
        const { search, role } = request.data;

        const query = this.queryBuilder.getInstance();
        query.sorting([{ id: "full_name", desc: false }]).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/users/base", {
                    params: {
                        ...query.build(),
                        ...(role ? { role } : {}),
                    },
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
