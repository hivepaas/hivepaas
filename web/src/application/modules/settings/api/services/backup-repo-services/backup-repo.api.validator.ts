import { type AxiosResponse } from "axios";
import { z } from "zod";
import type { SettingBackupRepo } from "~/settings/domain";
import { BackupRepoSettingEntitySchema } from "~/settings/module-shared/schemas";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    BackupRepo_DeleteOne_Res,
    BackupRepo_FindManyPaginated_Res,
    BackupRepo_FindOneById_Res,
    BackupRepo_UpdateStatus_Res,
} from "./backup-repo.api.contracts";

const FindManyPaginatedSchema = z.object({
    data: z.array(BackupRepoSettingEntitySchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: BackupRepoSettingEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class BackupRepoApiValidator {
    findManyPaginated = (response: AxiosResponse): BackupRepo_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data: data as unknown as SettingBackupRepo[], meta };
    };

    findOneById = (response: AxiosResponse): BackupRepo_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data: data as unknown as SettingBackupRepo, meta };
    };

    updateStatus = (response: AxiosResponse): BackupRepo_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: AxiosResponse): BackupRepo_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
