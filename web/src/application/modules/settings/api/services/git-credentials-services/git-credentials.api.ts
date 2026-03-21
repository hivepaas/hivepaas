import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { GitCredentialsApiValidator } from "./git-credentials.api.validator";
import type {
    GitCredentials_FindManyPaginated_Req,
    GitCredentials_FindManyPaginated_Res,
} from "./git-credentials.api.contracts";

export class GitCredentialsApi extends BaseApi {
    public constructor(private readonly validator: GitCredentialsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: GitCredentials_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<GitCredentials_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/git-credentials", {
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
