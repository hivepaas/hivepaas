import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    ProjectAppEnvVars_FindOne_Req,
    ProjectAppEnvVars_FindOne_Res,
    ProjectAppEnvVars_UpdateOne_Req,
    ProjectAppEnvVars_UpdateOne_Res,
} from "./project-app-env-vars.api.contracts";
import type { ProjectAppEnvVarsApiValidator } from "./project-app-env-vars.api.validator";

export class ProjectAppEnvVarsApi extends BaseApi {
    public constructor(private readonly validator: ProjectAppEnvVarsApiValidator) {
        super();
    }

    /**
     * Find one project app env vars
     */
    async findOne(
        request: ProjectAppEnvVars_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_FindOne_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/env-vars`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update project app env vars
     */
    async updateOne(
        request: ProjectAppEnvVars_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_UpdateOne_Res, Error>> {
        const { projectID, appID, updateVer, buildtime, runtime } = request.data;

        const json = {
            updateVer,
            buildtimeEnvVars: JsonTransformer.array({
                data: buildtime,
            }),
            runtimeEnvVars: JsonTransformer.array({
                data: runtime,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/env-vars`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
