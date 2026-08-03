import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    TraefikConfigOptions_FindOne_Req,
    TraefikConfigOptions_FindOne_Res,
    TraefikConfigOptions_UpdateOne_Req,
    TraefikConfigOptions_UpdateOne_Res,
} from "./traefik-config-options.api.contracts";
import type { TraefikConfigOptionsApiValidator } from "./traefik-config-options.api.validator";

export class TraefikConfigOptionsApi extends BaseApi {
    public constructor(private readonly validator: TraefikConfigOptionsApiValidator) {
        super();
    }

    async findOne(
        _request: TraefikConfigOptions_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikConfigOptions_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/traefik/config-options", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: TraefikConfigOptions_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikConfigOptions_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/traefik/config-options", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
