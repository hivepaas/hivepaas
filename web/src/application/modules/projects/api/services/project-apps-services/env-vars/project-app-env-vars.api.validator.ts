import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectAppEnvVars_Compute_Res, ProjectAppEnvVars_FindOne_Res } from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project buildtime env var schema
 */
const ProjectBuildtimeEnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().nullish(),
    isSystem: z.boolean().nullish(),
    isReadOnly: z.boolean().nullish(),
});

/**
 * Project runtime env var schema
 */
const ProjectRuntimeEnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().nullish(),
    isSystem: z.boolean().nullish(),
    isReadOnly: z.boolean().nullish(),
});

/**
 * Project shared env var schema
 */
const ProjectSharedEnvVarSchema = ProjectRuntimeEnvVarSchema;

function toDomainEnvVar(envVar: z.infer<typeof ProjectBuildtimeEnvVarSchema>) {
    return {
        key: envVar.key,
        value: envVar.value,
        isLiteral: envVar.isLiteral ?? false,
        isSystem: envVar.isSystem ?? false,
        isReadOnly: envVar.isReadOnly ?? false,
    };
}

/**
 * Project env var schema
 */
const ProjectEnvVarSchema = z.object({
    buildtimeEnvVars: z.array(ProjectBuildtimeEnvVarSchema),
    runtimeEnvVars: z.array(ProjectRuntimeEnvVarSchema),
    sharedEnvVars: z.array(ProjectSharedEnvVarSchema).nullish(),
    inheritedBuildtimeEnvVars: z.array(ProjectBuildtimeEnvVarSchema).nullish(),
    inheritedRuntimeEnvVars: z.array(ProjectRuntimeEnvVarSchema).nullish(),
    updateVer: z.number(),
});

/**
 * Find one project app env vars API response schema
 */
const FindOneSchema = z.object({
    data: ProjectEnvVarSchema.nullable(),
    meta: BaseMetaApiSchema.nullable(),
});

const ComputedEnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
});

/**
 * Compute project app env vars API response schema
 */
const ComputeSchema = z.object({
    data: z.array(ComputedEnvVarSchema).nullable(),
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
                buildtime: data ? data.buildtimeEnvVars.map(toDomainEnvVar) : [],
                runtime: data ? data.runtimeEnvVars.map(toDomainEnvVar) : [],
                shared: data?.sharedEnvVars ? data.sharedEnvVars.map(toDomainEnvVar) : [],
                inheritedBuildtimeEnvVars: data?.inheritedBuildtimeEnvVars
                    ? data.inheritedBuildtimeEnvVars.map(toDomainEnvVar)
                    : [],
                inheritedRuntimeEnvVars: data?.inheritedRuntimeEnvVars
                    ? data.inheritedRuntimeEnvVars.map(toDomainEnvVar)
                    : [],
                updateVer: data?.updateVer ?? 0,
            },
            meta,
        };
    };

    /**
     * Validate and transform compute project app env vars API response
     */
    compute = (response: AxiosResponse): ProjectAppEnvVars_Compute_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: ComputeSchema,
        });

        return {
            data: data ?? [],
            meta,
        };
    };
}
