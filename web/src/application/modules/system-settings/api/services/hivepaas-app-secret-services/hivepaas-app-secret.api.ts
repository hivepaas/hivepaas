import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSAppSecret_UpdateOne_Req,
    HivePaaSAppSecret_UpdateOne_Res,
} from "./hivepaas-app-secret.api.contracts";
import type { HivePaaSAppSecretApiValidator } from "./hivepaas-app-secret.api.validator";

export class HivePaaSAppSecretApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSAppSecretApiValidator) {
        super();
    }

    async updateOne(
        request: HivePaaSAppSecret_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSAppSecret_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/app-secret", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
