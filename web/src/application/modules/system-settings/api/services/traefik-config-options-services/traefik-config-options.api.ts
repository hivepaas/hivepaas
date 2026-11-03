import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    TraefikConfigOptions_ConfirmChange_Req,
    TraefikConfigOptions_ConfirmChange_Res,
    TraefikConfigOptions_FindOne_Req,
    TraefikConfigOptions_FindOne_Res,
    TraefikConfigOptions_RevertChange_Req,
    TraefikConfigOptions_RevertChange_Res,
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

    /**
     * Vouches for the change on trial, which is what stops it being undone.
     *
     * The proof is the request itself. Every route into HivePaaS goes through
     * Traefik, so this call can only arrive by having been served by the command
     * line it is confirming - which is the one thing Traefik's own healthcheck
     * cannot tell anybody.
     */
    async confirmChange(
        request: TraefikConfigOptions_ConfirmChange_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikConfigOptions_ConfirmChange_Res, Error>> {
        const { changeId } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/system/traefik/config-options/confirm", { changeId }, { signal })).pipe(
                map(this.validator.confirmChange),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async revertChange(
        request: TraefikConfigOptions_RevertChange_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikConfigOptions_RevertChange_Res, Error>> {
        const { changeId } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/system/traefik/config-options/revert", { changeId }, { signal })).pipe(
                map(this.validator.revertChange),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
