import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppDataFiles_CreateOne_Res,
    AppDataFiles_FindManyPaginated_Res,
    AppDataFiles_GetDownloadUrl_Res,
    AppDataFiles_UploadLocal_Res,
} from "~/projects/api/services/project-apps-services/data-files";
import { AppDataFileStorageType } from "~/projects/domain";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const AppDataFileSchema = z.object({
    id: z.string(),
    type: z.string(),
    kind: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    key: z
        .string()
        .nullish()
        .transform(value => value ?? ""),
    status: z.string(),
    name: z.string(),
    path: z.string(),
    bucket: z
        .string()
        .nullish()
        .transform(value => value ?? undefined),
    mimetype: z.string(),
    sizeBytes: z.number(),
    storageType: z.nativeEnum(AppDataFileStorageType),
    storage: z
        .object({
            name: z.string(),
        })
        .nullish()
        .transform(value => value ?? null),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppDataFileSchema),
    meta: PagingMetaApiSchema,
});

const GetDownloadUrlSchema = z.object({
    data: z.object({
        url: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const UploadLocalSchema = z.object({
    data: z.array(AppDataFileSchema),
    meta: BaseMetaApiSchema.nullish(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class AppDataFilesApiValidator {
    findManyPaginated = (response: AxiosResponse): AppDataFiles_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data, meta };
    };

    getDownloadUrl = (response: AxiosResponse): AppDataFiles_GetDownloadUrl_Res => {
        return parseApiResponse({
            response,
            schema: GetDownloadUrlSchema,
        });
    };

    uploadLocal = (response: AxiosResponse): AppDataFiles_UploadLocal_Res => {
        const parsed = parseApiResponse({
            response,
            schema: UploadLocalSchema,
        });
        return { data: { files: parsed.data }, meta: parsed.meta };
    };

    createOne = (response: AxiosResponse): AppDataFiles_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };
}
