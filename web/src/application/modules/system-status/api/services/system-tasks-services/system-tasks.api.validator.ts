import { type AxiosResponse } from "axios";
import { z } from "zod";
import { SystemTaskPriority, SystemTaskStatus } from "~/system-status/domain";

import { BaseMetaApiSchema, PagingMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemTasks_Cancel_Res,
    SystemTasks_FindManyPaginated_Res,
    SystemTasks_FindOneById_Res,
} from "./system-tasks.api.contracts";

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

const SystemTaskStatusSchema = z.enum([
    SystemTaskStatus.NotStarted,
    SystemTaskStatus.InProgress,
    SystemTaskStatus.Canceled,
    SystemTaskStatus.Done,
    SystemTaskStatus.Failed,
]);

const SystemTaskPrioritySchema = z
    .string()
    .optional()
    .default(SystemTaskPriority.Default)
    .transform(value =>
        Object.values(SystemTaskPriority).includes(value as SystemTaskPriority)
            ? (value as SystemTaskPriority)
            : SystemTaskPriority.Default,
    );

const SystemTaskConfigSchema = z.preprocess(
    value => value ?? {},
    z.object({
        priority: SystemTaskPrioritySchema,
        maxRetry: z.number().optional().default(0),
        retry: z.number().optional().default(0),
        retryDelay: z.string().optional().default(""),
        timeout: z.string().optional().default(""),
        controlDisabled: z.boolean().optional().default(false),
    }),
);

const SystemTaskTargetJobSchema = z
    .object({
        id: z.string().optional().default(""),
        type: z.string().optional().default(""),
        kind: z.string().optional().default(""),
        name: z.string().optional().default(""),
        status: z.string().optional().default(""),
    })
    .nullish()
    .transform(value => value ?? undefined);

const SystemTaskSchema = z.object({
    id: z.string(),
    type: z.string().optional().default(""),
    status: SystemTaskStatusSchema,
    config: SystemTaskConfigSchema,
    targetJob: SystemTaskTargetJobSchema,
    lastError: z.string().optional().default(""),
    updateVer: z.number().optional().default(0),
    runAt: NullableDateSchema,
    retryAt: NullableDateSchema,
    startedAt: NullableDateSchema,
    endedAt: NullableDateSchema,
    createdAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
});

const FindManyPaginatedSchema = z.object({
    data: z.array(SystemTaskSchema),
    meta: PagingMetaApiSchema,
});

const FindOneByIdSchema = z.object({
    data: SystemTaskSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const CancelSchema = z.object({
    data: z.object({
        canceled: z.boolean(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemTasksApiValidator {
    findManyPaginated = (response: AxiosResponse): SystemTasks_FindManyPaginated_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindManyPaginatedSchema });
        return { data, meta };
    };

    findOneById = (response: AxiosResponse): SystemTasks_FindOneById_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneByIdSchema });
        return { data, meta };
    };

    cancel = (response: AxiosResponse): SystemTasks_Cancel_Res => {
        const { data, meta } = parseApiResponse({ response, schema: CancelSchema });
        return { data, meta };
    };
}
