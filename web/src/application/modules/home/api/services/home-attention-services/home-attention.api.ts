import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { HomeAttention_FindAll_Req, HomeAttention_FindAll_Res } from "./home-attention.api.contracts";
import type { HomeAttentionApiValidator } from "./home-attention.api.validator";

export class HomeAttentionApi extends BaseApi {
    public constructor(private readonly validator: HomeAttentionApiValidator) {
        super();
    }

    async findAll(
        _request: HomeAttention_FindAll_Req,
        signal?: AbortSignal,
    ): Promise<Result<HomeAttention_FindAll_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/home/attention", { signal })).pipe(
                map(this.validator.findAll),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
