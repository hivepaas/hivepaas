import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectAppEnvVars_FindOne_Res } from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project buildtime env var schema
 */
const ProjectBuildtimeEnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().nullish(),
});

/**
 * Project runtime env var schema
 */
const ProjectRuntimeEnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().nullish(),
});

/**
 * Project env var schema
 */
const ProjectEnvVarSchema = z.object({
    buildtimeEnvVars: z.array(ProjectBuildtimeEnvVarSchema),
    runtimeEnvVars: z.array(ProjectRuntimeEnvVarSchema),
    updateVer: z.number(),
});

/**
 * Find one project app env vars API response schema
 */
const FindOneSchema = z.object({
    data: ProjectEnvVarSchema.nullable(),
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectAppEnvVarsApiValidator {
    /**
     * Validate and transform find one project app env vars API response
     */
    findOne = (response: AxiosResponse): ProjectAppEnvVars_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneSchema,
        });

        return {
            data: {
                buildtime: data
                    ? data.buildtimeEnvVars.map(envVar => ({
                          key: envVar.key,
                          value: envVar.value,
                          isLiteral: envVar.isLiteral ?? false,
                      }))
                    : [],
                runtime: data
                    ? data.runtimeEnvVars.map(envVar => ({
                          key: envVar.key,
                          value: envVar.value,
                          isLiteral: envVar.isLiteral ?? false,
                      }))
                    : [],
                updateVer: data?.updateVer ?? 0,
            },
            meta,
        };
    };
}
