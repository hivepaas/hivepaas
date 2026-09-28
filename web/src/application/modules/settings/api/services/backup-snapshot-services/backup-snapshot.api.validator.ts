import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    BackupSnapshot_DeleteOne_Res,
    BackupSnapshot_FindManyPaginated_Res,
    BackupSnapshot_FindOneById_Res,
} from "./backup-snapshot.api.contracts";

const RepoRefSchema = z.object({
    id: z.string(),
    name: z.string().optional().default(""),
    status: z.string().optional().default(""),
    scope: z.string().optional(),
});

const BackupSnapshotSchema = z.object({
    id: z.string(),
    snapshotId: z.string().optional().default(""),
    shortId: z.string().optional().default(""),
    time: z.coerce.date(),
    sizeBytes: z.number().optional().default(0),
    description: z.string().optional().default(""),
    paths: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    hostname: z.string().optional().default(""),
    tags: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    source: z.string().optional().default(""),
    repo: RepoRefSchema,
    app: z
        .object({
            id: z.string(),
            name: z.string().optional().default(""),
            env: z.string().optional().default(""),
            deleted: z.boolean().optional().default(false),
        })
        .nullish()
        .transform(value => value ?? undefined),
    job: z
        .object({
            id: z.string(),
            name: z.string().optional().default(""),
            deleted: z.boolean().optional().default(false),
        })
        .nullish()
        .transform(value => value ?? undefined),
    runId: z.string().optional().default(""),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(BackupSnapshotSchema),
    meta: PagingMetaApiSchema,
    repos: z
        .array(RepoRefSchema)
        .nullish()
        .transform(value => value ?? []),
});

const FindOneByIdSchema = z.object({
    data: BackupSnapshotSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class BackupSnapshotApiValidator {
    findManyPaginated = (response: AxiosResponse): BackupSnapshot_FindManyPaginated_Res => {
        const { data, meta, repos } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta, repos };
    };

    findOneById = (response: AxiosResponse): BackupSnapshot_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    deleteOne = (response: AxiosResponse): BackupSnapshot_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
