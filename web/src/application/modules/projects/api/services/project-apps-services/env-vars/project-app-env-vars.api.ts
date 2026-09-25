import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    EnvVarWireItem,
    ProjectAppEnvVars_Compute_Req,
    ProjectAppEnvVars_Compute_Res,
    ProjectAppEnvVars_FindLinkSuggestions_Req,
    ProjectAppEnvVars_FindLinkSuggestions_Res,
    ProjectAppEnvVars_FindLinkTargets_Req,
    ProjectAppEnvVars_FindLinkTargets_Res,
    ProjectAppEnvVars_FindOne_Req,
    ProjectAppEnvVars_FindOne_Res,
    ProjectAppEnvVars_UpdateOne_Req,
    ProjectAppEnvVars_UpdateOne_Res,
} from "./project-app-env-vars.api.contracts";
import type { ProjectAppEnvVarsApiValidator } from "./project-app-env-vars.api.validator";

function toEnvVarWire(envVars: { key: string; value: string; isLiteral: boolean }[]): EnvVarWireItem[] {
    return envVars.map(({ key, value, isLiteral }) => ({ key, value, isLiteral }));
}

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
        const { projectID, env, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/env-vars`, {
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
        const { projectID, env, appID, updateVer, buildtime, runtime, shared } = request.data;

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
            sharedEnvVars: JsonTransformer.array({
                data: shared,
                some: toEnvVarWire,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/${env}/apps/${appID}/env-vars`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Compute project app env vars
     */
    async compute(
        request: ProjectAppEnvVars_Compute_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_Compute_Res, Error>> {
        const { projectID, env, appID, buildtimeEnvVars, runtimeEnvVars, sharedEnvVars } = request.data;

        const json = {
            buildtimeEnvVars: JsonTransformer.array({
                data: buildtimeEnvVars,
                some: toEnvVarWire,
            }),
            runtimeEnvVars: JsonTransformer.array({
                data: runtimeEnvVars,
                some: toEnvVarWire,
            }),
            sharedEnvVars: JsonTransformer.array({
                data: sharedEnvVars,
                some: toEnvVarWire,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/${env}/apps/${appID}/env-vars/compute`, json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.compute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findLinkTargets(
        request: ProjectAppEnvVars_FindLinkTargets_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_FindLinkTargets_Res, Error>> {
        const { projectID, env, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/env-vars/link-targets`, { signal }),
            ).pipe(
                map(this.validator.findLinkTargets),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findLinkSuggestions(
        request: ProjectAppEnvVars_FindLinkSuggestions_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAppEnvVars_FindLinkSuggestions_Res, Error>> {
        const { projectID, env, appID, targetAppID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/env-vars/link-suggestions`, {
                    params: { targetAppId: targetAppID },
                    signal,
                }),
            ).pipe(
                map(this.validator.findLinkSuggestions),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
