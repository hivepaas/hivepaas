import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectEnvVarsApiValidator,
    ProjectEnvVars_FindOne_Req,
    ProjectEnvVars_FindOne_Res,
    ProjectEnvVars_UpdateOne_Req,
    ProjectEnvVars_UpdateOne_Res,
} from "~/projects/api/services/projects-services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProjectEnvVarsApi extends BaseApi {
    public constructor(private readonly validator: ProjectEnvVarsApiValidator) {
        super();
    }

    /**
     * Find one project env vars
     */
    async findOne(
        request: ProjectEnvVars_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEnvVars_FindOne_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/env-vars`, {
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
     * Update project env vars
     */
    async updateOne(
        request: ProjectEnvVars_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEnvVars_UpdateOne_Res, Error>> {
        const { projectID, updateVer, buildtime, runtime } = request.data;

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
                this.client.v1.put(`/projects/${projectID}/env-vars`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
