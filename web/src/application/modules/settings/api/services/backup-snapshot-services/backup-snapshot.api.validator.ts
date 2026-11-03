import { z } from "zod";
import { APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE } from "~/projects/domain";
import { EAppScheduledJobArgSeparator } from "~/projects/module-shared/enums";

import { type ApiHttpResponse, BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    BackupSnapshot_DeleteOne_Res,
    BackupSnapshot_FindEntries_Res,
    BackupSnapshot_FindManyPaginated_Res,
    BackupSnapshot_FindOneById_Res,
    BackupSnapshot_Restore_Res,
} from "./backup-snapshot.api.contracts";

const RepoRefSchema = z.object({
    id: z.string(),
    name: z.string().optional().default(""),
    status: z.string().optional().default(""),
    scope: z.string().optional(),
});

/** A job's command, as the job's form edits it. */
const CommandSchema = z.object({
    runInShell: z.string().optional().default(""),
    command: z.string().optional().default(""),
    script: z.string().optional().default(""),
    workingDir: z.string().optional().default(""),
    consoleSize: z
        .object({ width: z.number(), height: z.number() })
        .nullish()
        .transform(value => value ?? { ...APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE }),
    tty: z.boolean().optional().default(false),
    envVars: z
        .array(z.object({ key: z.string(), value: z.string(), isLiteral: z.boolean().optional().default(false) }))
        .nullish()
        .transform(value => value ?? []),
    argGroups: z
        .array(
            z.object({
                enabled: z.boolean().optional().default(false),
                exportEnv: z.string().optional().default(""),
                separator: z.nativeEnum(EAppScheduledJobArgSeparator).catch(EAppScheduledJobArgSeparator.Whitespace),
                args: z
                    .array(
                        z.object({
                            use: z.boolean().optional().default(false),
                            name: z.string().optional().default(""),
                            value: z.string().optional().default(""),
                        }),
                    )
                    .nullish()
                    .transform(value => value ?? []),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
});

const EntrySchema = z.object({
    name: z.string(),
    dir: z.boolean().optional().default(false),
    sizeBytes: z.number().optional().default(0),
});

const FindEntriesSchema = z.object({
    data: z
        .array(EntrySchema)
        .nullish()
        .transform(value => value ?? []),
    meta: BaseMetaApiSchema.nullish(),
});

const RestoreSchema = z.object({
    data: z.object({ task: z.object({ id: z.string() }) }),
    meta: BaseMetaApiSchema.nullish(),
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
            projectId: z.string().optional().default(""),
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
            fileName: z.string().optional().default(""),
            restoreCommand: CommandSchema.nullish().transform(value => value ?? undefined),
            sourceVolumeId: z.string().optional().default(""),
            sourceVolumeSubpath: z.string().optional().default(""),
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
    findManyPaginated = (response: ApiHttpResponse): BackupSnapshot_FindManyPaginated_Res => {
        const { data, meta, repos } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta, repos };
    };

    findOneById = (response: ApiHttpResponse): BackupSnapshot_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    findEntries = (response: ApiHttpResponse): BackupSnapshot_FindEntries_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindEntriesSchema });
        return { data, meta };
    };

    restore = (response: ApiHttpResponse): BackupSnapshot_Restore_Res => {
        const { data, meta } = parseApiResponse({ response, schema: RestoreSchema });
        return { data: { taskId: data.task.id }, meta };
    };

    deleteOne = (response: ApiHttpResponse): BackupSnapshot_DeleteOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
