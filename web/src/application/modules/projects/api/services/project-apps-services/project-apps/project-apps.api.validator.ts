import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ProjectApps_CreateOne_Res,
    ProjectApps_Deploy_Res,
    ProjectApps_DetectPhoto_Res,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Res,
} from "~/projects/api/services";

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

const DetectPhotoSchema = z.object({
    data: z.object({
        url: z.string(),
    }),
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

    detectPhoto = (response: AxiosResponse): ProjectApps_DetectPhoto_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: DetectPhotoSchema,
        });

        return {
            data,
            meta,
        };
    };
}
