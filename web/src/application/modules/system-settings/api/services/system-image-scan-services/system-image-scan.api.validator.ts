import { z } from "zod";
import { AppImageScanSchema, ScannedImageSchema, VulnSummarySchema } from "~/projects/api/services";
import { ImageScanAppSchema, SystemImageScanSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    SystemImageScan_Execute_Res,
    SystemImageScan_FindApps_Res,
    SystemImageScan_FindComponent_Res,
    SystemImageScan_FindComponents_Res,
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

const ComponentRefSchema = z.object({ key: z.string(), name: z.string() });

const FindComponentsSchema = z.object({
    data: z.object({
        components: z
            .array(
                z.object({
                    component: ComponentRefSchema,
                    scannedAt: z.coerce.date(),
                    error: z
                        .string()
                        .nullish()
                        .transform(value => value ?? ""),
                    summary: VulnSummarySchema,
                    images: z
                        .array(ScannedImageSchema)
                        .nullish()
                        .transform(value => value ?? []),
                }),
            )
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const FindComponentSchema = z.object({
    data: z.object({ component: ComponentRefSchema, scan: AppImageScanSchema }),
    meta: BaseMetaApiSchema.nullish(),
});

export class SystemImageScanApiValidator {
    findComponents = (response: ApiHttpResponse): SystemImageScan_FindComponents_Res => {
        return parseApiResponse({ response, schema: FindComponentsSchema });
    };

    findComponent = (response: ApiHttpResponse): SystemImageScan_FindComponent_Res => {
        return parseApiResponse({ response, schema: FindComponentSchema });
    };

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
