import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppLogs_GetHistory_Res, AppLogs_GetInfo_Res, AppLogs_GetLogs_Res } from "~/projects/api/services";

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
}
