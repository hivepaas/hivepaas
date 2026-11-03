import { z } from "zod";
import type {
    ProjectConfigFiles_CreateOne_Res,
    ProjectConfigFiles_FindManyPaginated_Res,
    ProjectConfigFiles_FindOneById_Res,
} from "~/projects/api/services/projects-services";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * Project config file schema
 */
const ProjectConfigFileSchema = z.object({
    id: z.string(),
    name: z.string(),
    content: z.string().optional().default(""),
    base64: z.boolean().optional().default(false),
    inheritable: z.boolean().optional().default(false),
    inherited: z.boolean().optional().default(false),
    status: z.string(),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(ProjectConfigFileSchema),
    meta: PagingMetaApiSchema,
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const FindOneByIdSchema = z.object({
    data: ProjectConfigFileSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class ProjectConfigFilesApiValidator {
    findManyPaginated = (response: ApiHttpResponse): ProjectConfigFiles_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    createOne = (response: ApiHttpResponse): ProjectConfigFiles_CreateOne_Res => {
        return parseApiResponse({ response, schema: CreateOneSchema });
    };

    findOneById = (response: ApiHttpResponse): ProjectConfigFiles_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };
}
