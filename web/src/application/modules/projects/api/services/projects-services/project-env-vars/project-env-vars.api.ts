import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    ProjectEnvVars_Compute_Req,
    ProjectEnvVars_Compute_Res,
    ProjectEnvVars_FindOne_Req,
    ProjectEnvVars_FindOne_Res,
    ProjectEnvVars_UpdateOne_Req,
    ProjectEnvVars_UpdateOne_Res,
} from "./project-env-vars.api.contracts";
import type { ProjectEnvVarsApiValidator } from "./project-env-vars.api.validator";

function toEnvVarWire(envVars: { key: string; value: string; isLiteral: boolean }[]) {
    return envVars.map(({ key, value, isLiteral }) => ({ key, value, isLiteral }));
}

function getEnvVarsBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/env-vars`;
    }

    return `/projects/${projectID}/env-vars`;
}

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
        const { projectID, env } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(getEnvVarsBasePath(projectID, env), {
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
        const { projectID, env, updateVer, buildtime, runtime } = request.data;

        const json = {
            updateVer,
            buildtimeEnvVars: JsonTransformer.array({
                data: buildtime,
                some: toEnvVarWire,
            }),
            runtimeEnvVars: JsonTransformer.array({
                data: runtime,
                some: toEnvVarWire,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(getEnvVarsBasePath(projectID, env), json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Compute project env vars
     */
    async compute(
        request: ProjectEnvVars_Compute_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEnvVars_Compute_Res, Error>> {
        const { projectID, env, buildtimeEnvVars, runtimeEnvVars } = request.data;

        const json = {
            buildtimeEnvVars: JsonTransformer.array({
                data: buildtimeEnvVars,
                some: toEnvVarWire,
            }),
            runtimeEnvVars: JsonTransformer.array({
                data: runtimeEnvVars,
                some: toEnvVarWire,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`${getEnvVarsBasePath(projectID, env)}/compute`, json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.compute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
