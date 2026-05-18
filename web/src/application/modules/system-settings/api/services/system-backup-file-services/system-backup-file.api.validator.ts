import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemBackupFileEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemBackupFile_FindManyPaginated_Res,
    SystemBackupFile_FindOneById_Res,
} from "./system-backup-file.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(SystemBackupFileEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: SystemBackupFileEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemBackupFileApiValidator {
    findManyPaginated = (response: AxiosResponse): SystemBackupFile_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): SystemBackupFile_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };
}
