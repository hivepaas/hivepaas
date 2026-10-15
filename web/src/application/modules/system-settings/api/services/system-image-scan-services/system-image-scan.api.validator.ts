import { z } from "zod";
import { ImageScanAppSchema, SystemImageScanSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemImageScan_Execute_Res,
    SystemImageScan_FindApps_Res,
    SystemImageScan_FindOne_Res,
    SystemImageScan_UpdateOne_Res,
} from "./system-image-scan.api.contracts";

const FindOneSchema = z.object({
    data: SystemImageScanSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const ExecuteSchema = z.object({
    data: z.object({
        task: z.object({
            id: z.string(),
        }),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const FindAppsSchema = z.object({
    data: z
        .array(ImageScanAppSchema)
        .nullish()
        .transform(value => value ?? []),
    meta: z.object({ page: z.object({ total: z.number().catch(0) }).nullish() }).nullish(),
});

export class SystemImageScanApiValidator {
    findOne = (response: ApiHttpResponse): SystemImageScan_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return { data, meta };
    };

    updateOne = (response: ApiHttpResponse): SystemImageScan_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    execute = (response: ApiHttpResponse): SystemImageScan_Execute_Res => {
        const { data, meta } = parseApiResponse({ response, schema: ExecuteSchema });
        return { data, meta };
    };

    findApps = (response: ApiHttpResponse): SystemImageScan_FindApps_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindAppsSchema });
        return { data, total: meta?.page?.total ?? data.length };
    };
}
