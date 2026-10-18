import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    AppManaged_CreateRestorePoint_Res,
    AppManaged_FindSettings_Res,
    AppManaged_FindStatus_Res,
    AppManaged_FindSwitch_Res,
    AppManaged_RefreshCredentials_Res,
    AppManaged_Restore_Res,
    AppManaged_Switch_Res,
} from "./app-managed.api.contracts";

const ObjectRefSchema = z
    .object({ id: z.string() })
    .nullish()
    .transform(value => (value?.id ? { id: value.id } : undefined));

const OptionalDateSchema = z.coerce
    .date()
    .nullish()
    .transform(value => value ?? undefined);

const StringListSchema = z
    .array(z.string())
    .nullish()
    .transform(value => value ?? []);

const PlacementSchema = z.enum(["off-site", "shared", "same-node", ""]).catch("").optional().default("");

const RestoreTargetSchema = z.object({
    kind: z.enum(["latest", "time", "point", "backup"]),
    time: OptionalDateSchema,
    point: z.string().optional().default(""),
    lsn: z.string().optional().default(""),
    backup: z.string().optional().default(""),
});

const SettingsSchema = z.object({
    updateVer: z.number(),
    engine: z.string(),
    engineApi: z.number(),
    storage: z
        .object({
            backupRepo: ObjectRefSchema,
            cloudStorage: ObjectRefSchema,
            keyAuth: ObjectRefSchema,
            prefix: z.string().optional().default(""),
            placement: PlacementSchema,
        })
        .nullish()
        .transform(value => value ?? undefined),
    archiveTimeout: z.string().optional().default(""),
    retainFull: z.number().optional().default(0),
    baseBackupSchedule: z.string().optional().default(""),
    restoreTestSchedule: z.string().optional().default(""),
    restoreTest: z
        .object({
            mode: z.enum(["container", "verifier"]).catch("container"),
            verifierApp: ObjectRefSchema,
            nodeId: z.string().optional().default(""),
            keyAuth: ObjectRefSchema,
            checkQuery: z.string().optional().default(""),
        })
        .nullish()
        .transform(value => value ?? undefined),
    dump: z
        .object({ repo: ObjectRefSchema, schedule: z.string().optional().default("") })
        .nullish()
        .transform(value => value ?? { repo: undefined, schedule: "" }),
    locks: z.object({
        image: z.string().optional().default(""),
        imageLine: z.string().optional().default(""),
        command: z.string().optional().default(""),
        mounts: StringListSchema,
        envKeys: StringListSchema,
        settings: StringListSchema,
    }),
    jobs: z
        .object({
            nightly: ObjectRefSchema,
            health: ObjectRefSchema,
            restoreTest: ObjectRefSchema,
            dump: ObjectRefSchema,
        })
        .nullish()
        .transform(value => value ?? undefined),
    restore: z
        .object({
            sourceApp: ObjectRefSchema,
            target: RestoreTargetSchema.nullish().transform(value => value ?? undefined),
            readyAt: OptionalDateSchema,
            switchedAt: OptionalDateSchema,
        })
        .nullish()
        .transform(value => value ?? undefined),
    switchedTo: z
        .object({ app: ObjectRefSchema, at: z.coerce.date() })
        .nullish()
        .transform(value => value ?? undefined),
});

const SwitchAppSchema = z.object({ id: z.string(), name: z.string(), key: z.string() });

const SwitchSchema = z.object({
    source: SwitchAppSchema,
    consumers: z
        .array(
            z.object({
                app: SwitchAppSchema,
                envVars: StringListSchema,
                previewClone: z.boolean().optional().default(false),
                deployRestorePoint: z.boolean().optional().default(false),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
    switchedAt: OptionalDateSchema,
});

const IdSchema = z.object({ id: z.string() });

const SwitchResultSchema = z.object({
    switched: z
        .array(IdSchema)
        .nullish()
        .transform(value => value ?? []),
    deployments: z
        .array(z.object({ app: IdSchema, deployment: IdSchema, task: IdSchema }))
        .nullish()
        .transform(value => value ?? []),
});

const RestorePointSchema = z.object({
    name: z.string(),
    lsn: z.string(),
    at: z.coerce.date(),
    by: ObjectRefSchema,
    note: z.string().optional().default(""),
    keepUntil: OptionalDateSchema,
    kept: z.boolean().optional().default(false),
    restorableUntil: OptionalDateSchema,
    origin: z.enum(["person", "deploy"]).catch("person").optional().default("person"),
});

const StatusSchema = z.object({
    engine: z.string(),
    engineApi: z.number(),
    archiving: z.object({
        state: z.enum(["ok", "failing", "unknown"]).catch("unknown"),
        checkedAt: OptionalDateSchema,
        lastArchivedAt: OptionalDateSchema,
        failedCount: z.number().optional().default(0),
        pgWalBytes: z.number().optional().default(0),
        stale: z.boolean().optional().default(false),
    }),
    restoreWindow: z
        .object({ from: z.coerce.date(), to: z.coerce.date() })
        .nullish()
        .transform(value => value ?? undefined),
    lastBackup: z
        .object({
            succeeded: z.boolean(),
            endedAt: z.coerce.date(),
            name: z.string().optional().default(""),
            at: OptionalDateSchema,
            bytes: z.number().optional().default(0),
        })
        .nullish()
        .transform(value => value ?? undefined),
    backups: z
        .array(z.object({ name: z.string(), at: z.coerce.date() }))
        .nullish()
        .transform(value => value ?? []),
    lastRestoreTest: z
        .object({
            passed: z.boolean(),
            endedAt: z.coerce.date(),
            durationSeconds: z.number().optional().default(0),
            restoredTo: z.string().optional().default(""),
        })
        .nullish()
        .transform(value => value ?? undefined),
    lastDump: z
        .object({
            succeeded: z.boolean(),
            endedAt: z.coerce.date(),
            snapshotId: z.string().optional().default(""),
            bytes: z.number().optional().default(0),
        })
        .nullish()
        .transform(value => value ?? undefined),
    placement: PlacementSchema,
    suggestions: StringListSchema,
    restorePoints: z
        .array(RestorePointSchema)
        .nullish()
        .transform(value => value ?? []),
    heldBack: z
        .object({ backups: z.number(), bytes: z.number() })
        .nullish()
        .transform(value => value ?? undefined),
});

const withMeta = <T extends z.ZodTypeAny>(data: T) => z.object({ data, meta: BaseMetaApiSchema.nullish() });

export class AppManagedApiValidator {
    findSettings = (response: ApiHttpResponse): AppManaged_FindSettings_Res => {
        return parseApiResponse({ response, schema: withMeta(SettingsSchema) });
    };

    findStatus = (response: ApiHttpResponse): AppManaged_FindStatus_Res => {
        return parseApiResponse({ response, schema: withMeta(StatusSchema) });
    };

    refreshCredentials = (response: ApiHttpResponse): AppManaged_RefreshCredentials_Res => {
        return parseApiResponse({ response, schema: withMeta(z.object({ changed: z.boolean() })) });
    };

    restore = (response: ApiHttpResponse): AppManaged_Restore_Res => {
        return parseApiResponse({
            response,
            schema: withMeta(z.object({ app: z.object({ id: z.string() }), task: z.object({ id: z.string() }) })),
        });
    };

    findSwitch = (response: ApiHttpResponse): AppManaged_FindSwitch_Res => {
        return parseApiResponse({ response, schema: withMeta(SwitchSchema) });
    };

    switchConsumers = (response: ApiHttpResponse): AppManaged_Switch_Res => {
        return parseApiResponse({ response, schema: withMeta(SwitchResultSchema) });
    };

    createRestorePoint = (response: ApiHttpResponse): AppManaged_CreateRestorePoint_Res => {
        return parseApiResponse({ response, schema: withMeta(RestorePointSchema) });
    };
}
