import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { TraefikRestart_Execute_Req, TraefikRestart_Execute_Res } from "./traefik-restart.api.contracts";
import type { TraefikRestartApiValidator } from "./traefik-restart.api.validator";

export class TraefikRestartApi extends BaseApi {
    public constructor(private readonly validator: TraefikRestartApiValidator) {
        super();
    }

    async execute(
        _request: TraefikRestart_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikRestart_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/traefik/restart", {}, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
