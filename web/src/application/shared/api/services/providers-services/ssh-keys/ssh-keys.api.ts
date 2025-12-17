import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { SshKeys_FindManyPaginated_Req, SshKeys_FindManyPaginated_Res } from "./ssh-keys.api.contracts";
import type { SshKeysApiValidator } from "./ssh-keys.api.validator";

export class SshKeysApi extends BaseApi {
    constructor(private readonly validator: SshKeysApiValidator) {
        super();
    }

    async findManyPaginated(
        req: SshKeys_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SshKeys_FindManyPaginated_Res, Error>> {
        const { pagination, sorting, search,status } = req.data;
        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search).filterBy({
            status: [status],
        });

        return lastValueFrom(
            from(
                this.client.v1.get("/providers/ssh-keys", {
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
