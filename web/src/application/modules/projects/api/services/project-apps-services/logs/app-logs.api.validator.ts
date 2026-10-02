import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    AppLogs_GetFunctionMetrics_Res,
    AppLogs_GetHistory_Res,
    AppLogs_GetInfo_Res,
    AppLogs_GetLogs_Res,
} from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const AppLogFrameSchema = z.object({
    type: z.enum(["in", "out", "err", "warn", "debug"]),
    data: z.string(),
    ts: z.coerce.date().nullable().catch(null),
});

const AppLogTaskSchema = z.object({
    id: z.string(),
});

const AppLogHistoryInfoSchema = z
    .object({
        available: z.boolean().catch(false),
        reason: z
            .enum(["disabled", "apps-not-collected", "no-query-endpoint", "driver-unreadable", "identity-missing"])
            .nullish()
            .transform(value => value ?? null)
            .catch(null),
        retention: z
            .string()
            .nullish()
            .transform(value => (value?.trim() ? value : undefined))
            .catch(undefined),
    })
    .nullish()
    // An older server says nothing; treat that as "not available" rather than
    // offering a tab that cannot load.
    .transform(value => value ?? { available: false, reason: "disabled" as const });

const GetInfoSchema = z.object({
    data: z.object({
        history: AppLogHistoryInfoSchema,
        enabled: z.boolean().optional().default(true),
        tasks: z
            .array(AppLogTaskSchema)
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const GetLogsSchema = z.object({
    data: z
        .object({
            logs: z
                .array(AppLogFrameSchema)
                .nullish()
                .transform(value => value ?? []),
        })
        .nullish()
        .transform(value => value ?? { logs: [] }),
    meta: BaseMetaApiSchema.nullable(),
});

const GetHistorySchema = z.object({
    data: z.object({
        logs: z
            .array(AppLogFrameSchema)
            .nullish()
            .transform(value => value ?? []),
        truncated: z.boolean().catch(false),
        nextEnd: z
            .string()
            .nullish()
            .transform(value => value ?? null),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const FunctionMetricsCountsSchema = z.object({
    calls: z.number().catch(0),
    failed: z.number().catch(0),
    errors5xx: z.number().catch(0),
    p50: z
        .number()
        .nullish()
        .transform(value => value ?? null),
    p95: z
        .number()
        .nullish()
        .transform(value => value ?? null),
    p99: z
        .number()
        .nullish()
        .transform(value => value ?? null),
});

const GetFunctionMetricsSchema = z.object({
    data: z.object({
        available: z.boolean().catch(false),
        reason: z
            .enum(["disabled", "apps-not-collected", "no-query-endpoint", "driver-unreadable", "identity-missing"])
            .nullish()
            .catch(null)
            .transform(value => value ?? null),
        range: z.enum(["1h", "6h", "24h", "7d"]).catch("24h"),
        start: z
            .string()
            .nullish()
            .transform(value => value ?? null),
        end: z
            .string()
            .nullish()
            .transform(value => value ?? null),
        stepSeconds: z.number().catch(0),
        clamped: z.boolean().catch(false),
        totals: FunctionMetricsCountsSchema.nullish().transform(value => value ?? null),
        byOutcome: z
            .record(z.number())
            .nullish()
            .transform(value => value ?? {}),
        series: z
            .array(FunctionMetricsCountsSchema.extend({ time: z.string() }))
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppLogsApiValidator {
    getInfo = (response: AxiosResponse): AppLogs_GetInfo_Res => {
        return parseApiResponse({
            response,
            schema: GetInfoSchema,
        });
    };

    getLogs = (response: AxiosResponse): AppLogs_GetLogs_Res => {
        const { data, meta } = parseApiResponse({
            response,
            schema: GetLogsSchema,
        });

        return {
            data: data.logs,
            meta,
        };
    };

    getHistory = (response: AxiosResponse): AppLogs_GetHistory_Res => {
        return parseApiResponse({
            response,
            schema: GetHistorySchema,
        });
    };

    getFunctionMetrics = (response: AxiosResponse): AppLogs_GetFunctionMetrics_Res => {
        return parseApiResponse({
            response,
            schema: GetFunctionMetricsSchema,
        });
    };
}
