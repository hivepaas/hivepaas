import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppLogs_GetInfo_Res, AppLogs_GetLogs_Res } from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const AppLogFrameSchema = z.object({
    type: z.enum(["in", "out", "err", "warn", "debug"]),
    data: z.string(),
    ts: z.coerce.date().nullable().catch(null),
});

const AppLogTaskSchema = z.object({
    id: z.string(),
});

const GetInfoSchema = z.object({
    data: z.object({
        enabled: z.boolean().optional().default(true),
        tasks: z.array(AppLogTaskSchema),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const GetLogsSchema = z.object({
    data: z.object({
        logs: z.array(AppLogFrameSchema),
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
}
