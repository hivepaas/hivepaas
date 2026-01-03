import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type Projects_CreateOne_Res,
    type Projects_FindManyPaginated_Res,
    type Projects_FindOneById_Res,
} from "~/projects/api/services/projects-services";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project schema
 */
const ProjectSchema = z.object({
    id: z.string(),
    name: z.string(),
    key: z.string(),
    status: z.nativeEnum(EProjectStatus),
    photo: z.string(),
    note: z.string(),
    tags: z.array(z.string()),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

/**
 * Find many projects paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create project API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Project user access schema
 */
const ProjectUserAccessSchema = z.object({
    id: z.string(),
    fullName: z.string(),
    access: z.object({
        read: z.boolean(),
        write: z.boolean(),
        delete: z.boolean(),
    }),
});

/**
 * Project details schema (includes userAccesses)
 */
const ProjectDetailsSchema = ProjectSchema.extend({
    userAccesses: z.array(ProjectUserAccessSchema),
});

/**
 * Find one project by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: ProjectDetailsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectsApiValidator {
    /**
     * Validate and transform find many projects paginated API response
     */
    findManyPaginated = (response: AxiosResponse): Projects_FindManyPaginated_Res => {
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
     * Validate and transform create project API response
     */
    createOne = (response: AxiosResponse): Projects_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    /**
     * Validate and transform find one project by id API response
     */
    findOneById = (response: AxiosResponse): Projects_FindOneById_Res => {
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
