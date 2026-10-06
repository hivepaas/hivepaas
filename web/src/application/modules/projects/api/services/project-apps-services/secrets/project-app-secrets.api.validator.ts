import { z } from "zod";
import type {
    AppSecrets_CreateOne_Res,
    AppSecrets_FindManyPaginated_Res,
    AppSecrets_FindOneById_Res,
    AppSecrets_GetDownloadToken_Res,
} from "~/projects/api/services/project-apps-services";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * App secret schema
 */
const AppSecretSchema = z.object({
    id: z.string(),
    name: z.string(),
    updateVer: z.number(),
    key: z.string(),
    value: z.string().optional(),
    secretMasked: z.boolean().optional(),
    base64: z.boolean().optional().default(false),
    inheritable: z.boolean().optional().default(false),
    type: z.string().optional().default("secret"),
    status: z.string(),
    inherited: z.boolean().optional().default(false),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
    expireAt: z.coerce.date().nullable().optional().default(null),
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

/**
 * Get app secret download token API response schema
 */
const GetDownloadTokenSchema = z.object({
    data: z.object({
        token: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppSecretsApiValidator {
    /**
     * Validate and transform find many app secrets paginated API response
     */
    findManyPaginated = (response: ApiHttpResponse): AppSecrets_FindManyPaginated_Res => {
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
    createOne = (response: ApiHttpResponse): AppSecrets_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    /**
     * Validate and transform find one app secret by id API response
     */
    findOneById = (response: ApiHttpResponse): AppSecrets_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return {
            data,
            meta,
        };
    };

    /**
     * Validate and transform get app secret download token API response
     */
    getDownloadToken = (response: ApiHttpResponse): AppSecrets_GetDownloadToken_Res => {
        return parseApiResponse({
            response,
            schema: GetDownloadTokenSchema,
        });
    };
}
