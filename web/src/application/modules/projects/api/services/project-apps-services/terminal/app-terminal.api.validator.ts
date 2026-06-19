import type { AxiosResponse } from "axios";
import { z } from "zod";
import type { AppTerminal_GetInfo_Res } from "~/projects/api/services";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

const GetInfoSchema = z.object({
    data: z.object({
        enabled: z.boolean().optional().default(true),
        supportedShells: z
            .array(z.string())
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppTerminalApiValidator {
    getInfo = (response: AxiosResponse): AppTerminal_GetInfo_Res => {
        return parseApiResponse({
            response,
            schema: GetInfoSchema,
        });
    };
}
