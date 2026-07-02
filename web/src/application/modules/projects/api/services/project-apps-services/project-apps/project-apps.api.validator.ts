import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ProjectApps_Copy_Res,
    ProjectApps_CreateOne_Res,
    ProjectApps_Deploy_Res,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Res,
    ProjectApps_PrepareCopy_Res,
} from "~/projects/api/services";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { ProjectAppDetailsSchema, ProjectAppSchema } from "./project-apps.api.schemas";

/**
 * Find many project apps paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectAppSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create project app API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const CopyToggleSchema = z.object({
    copy: z.boolean(),
});

const CopySslCertRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .nullish()
    .transform(value => value ?? null);

const CopyPreparedDomainSettingSchema = z.object({
    sourceDomain: z.string(),
    targetDomain: z.string(),
    sourceSslCert: CopySslCertRefSchema,
    targetSslCert: CopySslCertRefSchema,
});

const CopyHttpSettingsSchema = z.object({
    copy: z.boolean(),
    copyDomainSettings: z
        .array(CopyPreparedDomainSettingSchema)
        .nullish()
        .transform(value => value ?? []),
});

const PrepareCopySchema = z.object({
    data: z.object({
        sourceName: z.string(),
        targetName: z.string(),
        sourceEnv: z.string(),
        targetEnv: z.string(),
        sourceStatus: z.nativeEnum(EProjectAppStatus),
        targetStatus: z.nativeEnum(EProjectAppStatus),
        copyConfigFiles: CopyToggleSchema,
        copyDeploymentSettings: CopyToggleSchema,
        copyEnvVars: CopyToggleSchema,
        copyHealthChecks: CopyToggleSchema,
        copyHttpSettings: CopyHttpSettingsSchema,
        copySchedJobs: CopyToggleSchema,
        copySecrets: CopyToggleSchema,
        updateVer: z.number(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const CopySchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

/**
 * Deploy project app API response schema
 */
const DeploySchema = z.object({
    data: z.object({
        deploymentId: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Find one project app by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: ProjectAppDetailsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectAppsApiValidator {
    /**
     * Validate and transform find many project apps paginated API response
     */
    findManyPaginated = (response: AxiosResponse): ProjectApps_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return {
            data,
            meta,
        };
    };

    /**
     * Validate and transform create project app API response
     */
    createOne = (response: AxiosResponse): ProjectApps_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    prepareCopy = (response: AxiosResponse): ProjectApps_PrepareCopy_Res => {
        return parseApiResponse({
            response,
            schema: PrepareCopySchema,
        });
    };

    copy = (response: AxiosResponse): ProjectApps_Copy_Res => {
        return parseApiResponse({
            response,
            schema: CopySchema,
        });
    };

    /**
     * Validate and transform deploy project app API response
     */
    deploy = (response: AxiosResponse): ProjectApps_Deploy_Res => {
        return parseApiResponse({
            response,
            schema: DeploySchema,
        });
    };

    /**
     * Validate and transform find one project app by id API response
     */
    findOneById = (response: AxiosResponse): ProjectApps_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
            meta,
        };
    };
}
