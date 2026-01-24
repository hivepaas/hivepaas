import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    ProjectSecrets_CreateOne_Res,
    ProjectSecrets_FindManyPaginated_Res,
    ProjectSecrets_FindOneById_Res,
} from "~/projects/api/services/projects-services";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project secret schema
 */
const ProjectSecretSchema = z.object({
    id: z.string(),
    name: z.string(),
    updateVer: z.number(),
    key: z.string(),
    status: z.nativeEnum(EProjectSecretStatus),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

/**
 * Find many project secrets paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectSecretSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create project secret API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Find one project secret by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: ProjectSecretSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectSecretsApiValidator {
    /**
     * Validate and transform find many project secrets paginated API response
     */
    findManyPaginated = (response: AxiosResponse): ProjectSecrets_FindManyPaginated_Res => {
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
     * Validate and transform create project secret API response
     */
    createOne = (response: AxiosResponse): ProjectSecrets_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    /**
     * Validate and transform find one project secret by id API response
     */
    findOneById = (response: AxiosResponse): ProjectSecrets_FindOneById_Res => {
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
