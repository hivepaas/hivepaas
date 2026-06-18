import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppServiceTasksApiValidator,
    AppServiceTasks_FindMany_Req,
    AppServiceTasks_FindMany_Res,
} from "~/projects/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppServiceTasksApi extends BaseApi {
    public constructor(private readonly validator: AppServiceTasksApiValidator) {
        super();
    }

    async findMany(
        request: AppServiceTasks_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppServiceTasks_FindMany_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/service-tasks`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findMany),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
