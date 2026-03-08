import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppSecrets_CreateOne_Res,
    AppSecrets_FindManyPaginated_Res,
    AppSecrets_FindOneById_Res,
} from "~/projects/api/services/project-apps-services/project-app-secrets";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * App secret schema
 */
const AppSecretSchema = z.object({
    id: z.string(),
    name: z.string(),
    updateVer: z.number(),
    key: z.string(),
    status: z.nativeEnum(EProjectSecretStatus),
    inherited: z.boolean().optional().default(false),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

/**
 * Find many app secrets paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(AppSecretSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create app secret API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Find one app secret by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: AppSecretSchema,
    meta: BaseMetaApiSchema.nullable(),
});

export class AppSecretsApiValidator {
    /**
     * Validate and transform find many app secrets paginated API response
     */
    findManyPaginated = (response: AxiosResponse): AppSecrets_FindManyPaginated_Res => {
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
     * Validate and transform create app secret API response
     */
    createOne = (response: AxiosResponse): AppSecrets_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    /**
     * Validate and transform find one app secret by id API response
     */
    findOneById = (response: AxiosResponse): AppSecrets_FindOneById_Res => {
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
