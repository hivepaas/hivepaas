import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    TraefikServiceSettings_FindOne_Req,
    TraefikServiceSettings_FindOne_Res,
    TraefikServiceSettings_UpdateOne_Req,
    TraefikServiceSettings_UpdateOne_Res,
} from "./traefik-service-settings.api.contracts";
import type { TraefikServiceSettingsApiValidator } from "./traefik-service-settings.api.validator";

export class TraefikServiceSettingsApi extends BaseApi {
    public constructor(private readonly validator: TraefikServiceSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: TraefikServiceSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikServiceSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/traefik/service-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: TraefikServiceSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<TraefikServiceSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/traefik/service-settings", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
