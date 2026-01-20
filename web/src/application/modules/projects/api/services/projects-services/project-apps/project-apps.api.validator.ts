import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ProjectApps_CreateOne_Res,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Res,
} from "~/projects/api/services/projects-services";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project app schema
 */
const ProjectAppSchema = z.object({
    id: z.string(),
    name: z.string(),
    status: z.nativeEnum(EProjectAppStatus),
    note: z.string(),
    tags: z.array(z.string()),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

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

/**
 * Project app user access schema
 */
const ProjectAppUserAccessSchema = z.object({
    id: z.string(),
    fullName: z.string(),
    access: z.object({
        read: z.boolean(),
        write: z.boolean(),
        delete: z.boolean(),
    }),
});

/**
 * Project app stats schema
 */
const ProjectAppStatsSchema = z.object({
    runningTasks: z.number(),
    desiredTasks: z.number(),
    completedTasks: z.number(),
});

/**
 * Project app details schema
 */
const ProjectAppDetailsSchema = ProjectAppSchema.extend({
    key: z.string(),
    photo: z.string().nullable(),
    userAccesses: z.array(ProjectAppUserAccessSchema),
    updateVer: z.number(),
    stats: ProjectAppStatsSchema,
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
