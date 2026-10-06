import { z } from "zod";
import type { SettingBackupRepo } from "~/settings/domain";
import { BackupRepoSettingEntitySchema } from "~/settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    BackupRepo_Cleanup_Res,
    BackupRepo_CreateOne_Res,
    BackupRepo_DeleteOne_Res,
    BackupRepo_FindManyPaginated_Res,
    BackupRepo_FindOneById_Res,
    BackupRepo_Sync_Res,
    BackupRepo_UpdateOne_Res,
    BackupRepo_UpdatePassword_Res,
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

const CreateOneSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class BackupRepoApiValidator {
    findManyPaginated = (response: ApiHttpResponse): BackupRepo_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });

        return { data: data as unknown as SettingBackupRepo[], meta };
    };

    findOneById = (response: ApiHttpResponse): BackupRepo_FindOneById_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });

        return { data: data as unknown as SettingBackupRepo, meta };
    };

    createOne = (response: ApiHttpResponse): BackupRepo_CreateOne_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: CreateOneSchema,
        });

        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): BackupRepo_UpdateOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updateStatus = (response: ApiHttpResponse): BackupRepo_UpdateStatus_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    updatePassword = (response: ApiHttpResponse): BackupRepo_UpdatePassword_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    deleteOne = (response: ApiHttpResponse): BackupRepo_DeleteOne_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    cleanup = (response: ApiHttpResponse): BackupRepo_Cleanup_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };

    sync = (response: ApiHttpResponse): BackupRepo_Sync_Res => {
        parseApiResponse({
            response,
            schema: MetaOnlySchema,
        });

        return { data: { type: "success" } };
    };
}
