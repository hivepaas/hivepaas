import { type AxiosResponse } from "axios";
import { z } from "zod";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { McpSettings_FindOne_Res, McpSettings_UpdateOne_Res } from "./mcp-settings.api.contracts";

const FindOneSchema = z.object({
    data: z.object({
        enabled: z.boolean(),
        allowWrite: z.boolean(),
        updateVer: z.number(),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class McpSettingsApiValidator {
    findOne = (response: AxiosResponse): McpSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: AxiosResponse): McpSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
