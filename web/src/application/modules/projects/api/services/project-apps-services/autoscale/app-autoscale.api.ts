import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppAutoscale_FindOne_Req,
    AppAutoscale_FindOne_Res,
    AppAutoscale_UpdateOne_Req,
    AppAutoscale_UpdateOne_Res,
} from "./app-autoscale.api.contracts";
import type { AppAutoscaleApiValidator } from "./app-autoscale.api.validator";

export class AppAutoscaleApi extends BaseApi {
    constructor(private readonly validator: AppAutoscaleApiValidator) {
        super();
    }

    async findOne(
        req: AppAutoscale_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppAutoscale_FindOne_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/autoscale`, { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        req: AppAutoscale_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppAutoscale_UpdateOne_Res, Error>> {
        const { projectID, env, appID, payload } = req.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/autoscale`, payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
