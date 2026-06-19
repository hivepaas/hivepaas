import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppTerminalApiValidator,
    AppTerminal_GetInfo_Req,
    AppTerminal_GetInfo_Res,
} from "~/projects/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppTerminalApi extends BaseApi {
    public constructor(private readonly validator: AppTerminalApiValidator) {
        super();
    }

    async getInfo(
        request: AppTerminal_GetInfo_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppTerminal_GetInfo_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/terminal/info`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.getInfo),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
