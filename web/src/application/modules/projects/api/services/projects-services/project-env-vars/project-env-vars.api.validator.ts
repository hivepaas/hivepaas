import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { ProjectEnvVars_Compute_Res, ProjectEnvVars_FindOne_Res } from "~/projects/api/services/projects-services";

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
    inheritedBuildtimeEnvVars: z.array(ProjectBuildtimeEnvVarSchema).nullish(),
    buildtimeEnvVars: z.array(ProjectBuildtimeEnvVarSchema),
    inheritedRuntimeEnvVars: z.array(ProjectRuntimeEnvVarSchema).nullish(),
    runtimeEnvVars: z.array(ProjectRuntimeEnvVarSchema),
    updateVer: z.number(),
});

/**
 * Find one project env vars API response schema
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
 * Compute project env vars API response schema
 */
const ComputeSchema = z.object({
    data: z.array(ComputedEnvVarSchema).nullable(),
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectEnvVarsApiValidator {
    /**
     * Validate and transform find one project env vars API response
     */
    findOne = (response: AxiosResponse): ProjectEnvVars_FindOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneSchema,
        });

        return {
            data: {
                inheritedBuildtimeEnvVars: data?.inheritedBuildtimeEnvVars
                    ? data.inheritedBuildtimeEnvVars.map(toDomainEnvVar)
                    : [],
                buildtime: data ? data.buildtimeEnvVars.map(toDomainEnvVar) : [],
                inheritedRuntimeEnvVars: data?.inheritedRuntimeEnvVars
                    ? data.inheritedRuntimeEnvVars.map(toDomainEnvVar)
                    : [],
                runtime: data ? data.runtimeEnvVars.map(toDomainEnvVar) : [],
                updateVer: data?.updateVer ?? 0,
            },
            meta,
        };
    };

    /**
     * Validate and transform compute project env vars API response
     */
    compute = (response: AxiosResponse): ProjectEnvVars_Compute_Res => {
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
