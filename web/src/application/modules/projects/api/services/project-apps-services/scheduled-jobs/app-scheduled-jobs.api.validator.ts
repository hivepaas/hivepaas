import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppScheduledJobTasks_FindManyPaginated_Res,
    AppScheduledJobTasks_FindOneById_Res,
    AppScheduledJobTasks_GetLogs_Res,
    AppScheduledJobs_CreateOne_Res,
    AppScheduledJobs_FindManyPaginated_Res,
    AppScheduledJobs_FindOneById_Res,
    AppScheduledJobs_RunNow_Res,
} from "~/projects/api/services/project-apps-services";
import { APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE } from "~/projects/domain";
import {
    EAppScheduledJobArgSeparator,
    EAppScheduledJobTaskPriority,
    EAppScheduledJobTaskStatus,
    EAppScheduledJobType,
} from "~/projects/module-shared/enums";

import { ESettingStatus, ESettingType } from "@application/shared/enums";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const NamedRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .nullish()
    .transform(value => value ?? undefined);

const ScheduleSchema = z.object({
    cronExpr: z.string().optional().default(""),
    interval: z.string().optional().default(""),
    initialTime: z.coerce.date(),
});

const EnvVarSchema = z.object({
    key: z.string(),
    value: z.string(),
    isLiteral: z.boolean().optional().default(false),
});

const ArgSchema = z.object({
    use: z.boolean().optional().default(false),
    name: z.string().optional().default(""),
    value: z.string().optional().default(""),
});

const ArgGroupSchema = z.object({
    enabled: z.boolean().optional().default(false),
    exportEnv: z.string().optional().default(""),
    separator: z
        .string()
        .optional()
        .default(EAppScheduledJobArgSeparator.Whitespace)
        .transform(value =>
            value === EAppScheduledJobArgSeparator.Equal
                ? EAppScheduledJobArgSeparator.Equal
                : EAppScheduledJobArgSeparator.Whitespace,
        ),
    args: z
        .array(ArgSchema)
        .nullish()
        .transform(value => value ?? []),
});

const ConsoleSizeSchema = z
    .object({
        width: z.number(),
        height: z.number(),
    })
    .nullish()
    .transform(value => value ?? { ...APP_SCHEDULED_JOB_DEFAULT_CONSOLE_SIZE });

const CommandSchema = z
    .object({
        runInShell: z.string().optional().default(""),
        command: z.string().optional().default(""),
        script: z.string().optional().default(""),
        workingDir: z.string().optional().default(""),
        consoleSize: ConsoleSizeSchema,
        tty: z.boolean().optional().default(false),
        envVars: z
            .array(EnvVarSchema)
            .nullish()
            .transform(value => value ?? []),
        argGroups: z
            .array(ArgGroupSchema)
            .nullish()
            .transform(value => value ?? []),
    })
    .nullish()
    .transform(value => value ?? null);

const NotificationSchema = z
    .object({
        successUseDefault: z.boolean(),
        success: NamedRefSchema,
        failureUseDefault: z.boolean(),
        failure: NamedRefSchema,
    })
    .nullish()
    .transform(value => value ?? null);

const AppScheduledJobSchema = z.object({
    id: z.string(),
    type: z.nativeEnum(ESettingType),
    name: z.string(),
    kind: z.string().optional(),
    status: z.nativeEnum(ESettingStatus),
    inherited: z.boolean().optional().default(false),
    availableInProjects: z.boolean().optional().default(false),
    default: z.boolean().optional().default(false),
    updateVer: z.number(),
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
    expireAt: z.coerce.date().nullable().optional().default(null),
    jobType: z.nativeEnum(EAppScheduledJobType),
    schedule: ScheduleSchema,
    app: NamedRefSchema,
    priority: z.nativeEnum(EAppScheduledJobTaskPriority),
    maxRetry: z.number().optional().default(0),
    retryDelay: z.string().optional().default(""),
    timeout: z.string().optional().default(""),
    controlDisabled: z.boolean().optional().default(false),
    command: CommandSchema,
    notification: NotificationSchema,
    nextRuns: z.array(z.coerce.date()).optional().default([]),
});

const NullableDateSchema = z.preprocess(value => {
    if (value === null || value === undefined || value === "") {
        return null;
    }

    if (value instanceof Date) {
        return Number.isNaN(value.getTime()) ? null : value;
    }

    if (typeof value !== "string" && typeof value !== "number") {
        return null;
    }

    const date = new Date(value);

    return Number.isNaN(date.getTime()) ? null : date;
}, z.date().nullable());

const AppScheduledJobTaskPrioritySchema = z
    .string()
    .optional()
    .default(EAppScheduledJobTaskPriority.Default)
    .transform(value =>
        Object.values(EAppScheduledJobTaskPriority).includes(value as EAppScheduledJobTaskPriority)
            ? (value as EAppScheduledJobTaskPriority)
            : EAppScheduledJobTaskPriority.Default,
    );

const AppScheduledJobTaskConfigSchema = z.preprocess(
    value => value ?? {},
    z.object({
        priority: AppScheduledJobTaskPrioritySchema,
        maxRetry: z.number().optional().default(0),
        retry: z.number().optional().default(0),
        retryDelay: z.string().optional().default(""),
        timeout: z.string().optional().default(""),
        controlDisabled: z.boolean().optional().default(false),
    }),
);

const AppScheduledJobTaskSchema = z.object({
    id: z.string(),
    type: z.string().optional().default(""),
    status: z.nativeEnum(EAppScheduledJobTaskStatus),
    config: AppScheduledJobTaskConfigSchema,
    lastError: z.string().optional().default(""),
    updateVer: z.number(),
    runAt: NullableDateSchema,
    retryAt: NullableDateSchema,
    startedAt: NullableDateSchema,
    endedAt: NullableDateSchema,
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
});

const AppScheduledJobTaskLogFrameSchema = z.object({
    type: z.enum(["in", "out", "err", "warn", "debug"]),
    data: z.string(),
    ts: NullableDateSchema,
});

const FindManyPaginatedSchema = z.object({
    data: z.array(AppScheduledJobSchema),
    meta: PagingMetaApiSchema,
});

const FindTasksManyPaginatedSchema = z.object({
    data: z.array(AppScheduledJobTaskSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: AppScheduledJobSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const FindTaskByIdSchema = z.object({
    data: AppScheduledJobTaskSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const CreateOneSchema = z.object({
    data: z.object({
        id: z.string(),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const RunNowSchema = z.object({
    data: z.object({
        task: z.object({
            id: z.string(),
        }),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const GetTaskLogsSchema = z.object({
    data: z.object({
        logs: z.array(AppScheduledJobTaskLogFrameSchema),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppScheduledJobsApiValidator {
    findManyPaginated = (response: AxiosResponse): AppScheduledJobs_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindManyPaginatedSchema,
        });
    };

    findOneById = (response: AxiosResponse): AppScheduledJobs_FindOneById_Res => {
        return parseApiResponse({
            response,
            schema: FindOneByIdSchema,
        });
    };

    findTasksManyPaginated = (response: AxiosResponse): AppScheduledJobTasks_FindManyPaginated_Res => {
        return parseApiResponse({
            response,
            schema: FindTasksManyPaginatedSchema,
        });
    };

    findTaskById = (response: AxiosResponse): AppScheduledJobTasks_FindOneById_Res => {
        return parseApiResponse({
            response,
            schema: FindTaskByIdSchema,
        });
    };

    createOne = (response: AxiosResponse): AppScheduledJobs_CreateOne_Res => {
        return parseApiResponse({
            response,
            schema: CreateOneSchema,
        });
    };

    runNow = (response: AxiosResponse): AppScheduledJobs_RunNow_Res => {
        return parseApiResponse({
            response,
            schema: RunNowSchema,
        });
    };

    getTaskLogs = (response: AxiosResponse): AppScheduledJobTasks_GetLogs_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: GetTaskLogsSchema,
        });

        return {
            data: data.logs,
            meta,
        };
    };
}
