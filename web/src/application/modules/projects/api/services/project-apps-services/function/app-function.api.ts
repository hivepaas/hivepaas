import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import { type AppFunction_TestRun_Req, type AppFunction_TestRun_Res } from "./app-function.api.contracts";
import { type AppFunctionApiValidator } from "./app-function.api.validator";

export class AppFunctionApi extends BaseApi {
    constructor(private readonly validator: AppFunctionApiValidator) {
        super();
    }

    /**
     * Calls the function once with the code sent, on a build node. The first run
     * after the libraries change installs them, which takes minutes.
     */
    async testRun(req: AppFunction_TestRun_Req, signal?: AbortSignal): Promise<Result<AppFunction_TestRun_Res, Error>> {
        const { projectID, env, appID, files, request } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.post(
                    `/projects/${projectID}/${env}/apps/${appID}/function/test-run`,
                    { code: { files }, request },
                    { signal },
                ),
            ).pipe(
                map(this.validator.testRun),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
