import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import {
    type UsersApiValidator,
    type Users_DeleteOne_Req,
    type Users_DeleteOne_Res,
    type Users_FindManyPaginated_Req,
    type Users_FindManyPaginated_Res,
} from "~/users/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class UsersApi extends BaseApi {
    public constructor(private readonly validator: UsersApiValidator) {
        super();
    }

    /**
     * Find many users paginated
     */
    async findManyPaginated(
        request: Users_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Users_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/users", {
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

    /**
     * Delete a user
     */
    async deleteOne(request: Users_DeleteOne_Req): Promise<Result<Users_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/users/${id}`, {})).pipe(
                map(() => Ok({ data: { id } })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
