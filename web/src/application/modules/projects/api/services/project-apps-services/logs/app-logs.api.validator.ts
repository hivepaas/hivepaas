import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppLogs_GetInfo_Res, AppLogs_GetLogs_Res, AppLogs_GetToken_Res } from "~/projects/api/services";

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
        tasks: z.array(AppLogTaskSchema),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

const GetTokenSchema = z.object({
    data: z.object({
        token: z.string(),
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

    getToken = (response: AxiosResponse): AppLogs_GetToken_Res => {
        return parseApiResponse({
            response,
            schema: GetTokenSchema,
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
