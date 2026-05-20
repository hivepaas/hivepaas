import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppConfigFiles_CreateOne_Res,
    AppConfigFiles_FindManyPaginated_Res,
    AppConfigFiles_FindOneById_Res,
    AppConfigFiles_GetDownloadToken_Res,
} from "~/projects/api/services/project-apps-services";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * App config file schema
 */
const AppConfigFileSchema = z.object({
    id: z.string(),
    name: z.string(),
    content: z.string().optional().default(""),
    base64: z.boolean().optional().default(false),
    type: z.string().optional().default("config-file"),
    status: z.nativeEnum(EProjectSecretStatus),
    inherited: z.boolean().optional().default(false),
    swarmRef: z
        .object({
            file: z
                .object({
                    name: z.string(),
                    uid: z.string().optional().default(""),
                    gid: z.string().optional().default(""),
                    mode: z.union([z.string(), z.number()]).transform(value => String(value)),
                })
                .nullable()
                .optional()
                .default(null),
        })
        .nullable()
        .optional()
        .default(null),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
    expireAt: z.coerce.date().nullable().optional().default(null),
});

/**
 * Find many app config files paginated API response schema
 */
const FindManyPaginatedSchema = z.object({
    data: z.array(AppConfigFileSchema),
    meta: PagingMetaApiSchema,
});

/**
 * Create app config file API response schema
 */
const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Find one app config file by id API response schema
 */
const FindOneByIdSchema = z.object({
    data: AppConfigFileSchema,
    meta: BaseMetaApiSchema.nullable(),
});

/**
 * Get app config file download token API response schema
 */
const GetDownloadTokenSchema = z.object({
    data: z.object({
        token: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppConfigFilesApiValidator {
    /**
     * Validate and transform find many app config files paginated API response
     */
    findManyPaginated = (response: AxiosResponse): AppConfigFiles_FindManyPaginated_Res => {
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
     * Validate and transform create app config file API response
     */
    createOne = (response: AxiosResponse): AppConfigFiles_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    /**
     * Validate and transform find one app config file by id API response
     */
    findOneById = (response: AxiosResponse): AppConfigFiles_FindOneById_Res => {
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
     * Validate and transform get app config file download token API response
     */
    getDownloadToken = (response: AxiosResponse): AppConfigFiles_GetDownloadToken_Res => {
        return parseApiResponse({
            response,
            schema: GetDownloadTokenSchema,
        });
    };
}
