import { type AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppSettingMounts_CreateOne_Res,
    AppSettingMounts_FindManyPaginated_Res,
    AppSettingMounts_FindOneById_Res,
    AppSettingMounts_FindSources_Res,
} from "~/projects/api/services/project-apps-services";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

/**
 * App setting mount schema
 */
const AppSettingMountSchema = z.object({
    id: z.string(),
    name: z.string(),
    status: z.string(),
    inheritable: z.boolean().optional().default(false),
    updateVer: z.number(),
    source: z.object({
        id: z.string(),
        name: z.string().optional().default(""),
        type: z.string().optional().default(""),
        status: z.string().optional().default(""),
    }),
    files: z
        .array(
            z.object({
                part: z.string(),
                path: z.string(),
                uid: z.string().optional().default(""),
                gid: z.string().optional().default(""),
                mode: z.union([z.string(), z.number()]).transform(value => String(value)),
                secret: z.boolean().optional().default(false),
                gated: z.boolean().optional().default(false),
            }),
        )
        .nullable()
        .transform(value => value ?? []),
    state: z
        .object({
            reason: z.string().optional().default(""),
            mounted: z
                .array(z.string())
                .nullable()
                .optional()
                .transform(value => value ?? []),
        })
        .nullable()
        .optional()
        .default(null),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date().nullable(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppSettingMountSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AppSettingMountSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const FindSourcesSchema = z.object({
    data: z
        .array(
            z.object({
                type: z.string(),
                parts: z.array(
                    z.object({
                        name: z.string(),
                        required: z.boolean(),
                        secret: z.boolean().optional().default(false),
                        gated: z.boolean().optional().default(false),
                    }),
                ),
            }),
        )
        .nullable()
        .transform(value => value ?? []),
    mayMountSensitive: z.boolean(),
    meta: BaseMetaApiSchema.nullable().optional().default(null),
});

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppSettingMountsApiValidator {
    findManyPaginated = (response: AxiosResponse): AppSettingMounts_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): AppSettingMounts_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    findSources = (response: AxiosResponse): AppSettingMounts_FindSources_Res => {
        const { data, mayMountSensitive, meta } = parseApiResponse({ response, schema: FindSourcesSchema });
        return { data: { sources: data, mayMountSensitive }, meta };
    };

    createOne = (response: AxiosResponse): AppSettingMounts_CreateOne_Res => {
        return parseApiResponse({ response, schema: CreateOneSchema });
    };
}
