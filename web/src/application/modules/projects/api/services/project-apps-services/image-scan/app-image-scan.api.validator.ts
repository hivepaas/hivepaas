import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    AppImageScan_CreateIgnore_Res,
    AppImageScan_DeleteIgnore_Res,
    AppImageScan_FindOne_Res,
    AppImageScan_Scan_Res,
    ProjectImageScan_Find_Res,
    VulnSeverity,
} from "./app-image-scan.api.contracts";

const SEVERITIES: VulnSeverity[] = ["CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"];

const text = z
    .string()
    .nullish()
    .transform(value => value ?? "");

const date = z.coerce
    .date()
    .nullish()
    .transform(value => value ?? null);

const CountSchema = z.object({
    total: z.number().catch(0),
    fixable: z.number().catch(0),
});

export const VulnSummarySchema = z.object({
    critical: CountSchema,
    high: CountSchema,
    medium: CountSchema,
    low: CountSchema,
    unknown: CountSchema,
    ignored: z.number().catch(0),
});

export const ScannedImageSchema = z.object({
    imageId: z.string(),
    imageRef: text,
    platform: text,
    osFamily: text,
    osName: text,
    osEosl: z.boolean().catch(false),
    status: z.string(),
    source: text,
    error: text,
    scannedAt: z.coerce.date(),
    updatedAt: z.coerce.date(),
    dbUpdatedAt: date,
    trivyVersion: text,
    summary: VulnSummarySchema,
});

const FindingSchema = z.object({
    imageId: z.string(),
    vulnId: z.string(),
    pkgName: z.string(),
    pkgType: text,
    pkgPath: text,
    installedVersion: text,
    fixedVersion: text,
    severity: z
        .string()
        .transform(value => (SEVERITIES.includes(value as VulnSeverity) ? (value as VulnSeverity) : "UNKNOWN")),
    title: text,
    url: text,
    publishedAt: date,
});

const IgnoredSchema = FindingSchema.extend({
    ignore: z.object({
        id: z.string(),
        scope: z.string().transform(value => (value === "" ? "global" : value) as "app"),
        package: text,
        reason: text,
        setBy: text,
        expireAt: date,
        inheritable: z.boolean().catch(true),
    }),
});

export const AppImageScanSchema = z.object({
    enabled: z.boolean().catch(false),
    scannedAt: date,
    error: text,
    summary: VulnSummarySchema,
    images: z
        .array(ScannedImageSchema)
        .nullish()
        .transform(value => value ?? []),
    findings: z
        .array(FindingSchema)
        .nullish()
        .transform(value => value ?? []),
    ignored: z
        .array(IgnoredSchema)
        .nullish()
        .transform(value => value ?? []),
    task: z
        .object({ id: z.string(), status: z.string() })
        .nullish()
        .transform(value => value ?? null),
});

const ScannedAppSchema = z.object({
    app: z.object({ id: z.string(), name: z.string() }),
    project: z.object({ id: z.string(), name: z.string() }),
    projectEnv: z.object({ id: z.string(), key: text, name: z.string() }),
    scannedAt: z.coerce.date(),
    error: text,
    summary: VulnSummarySchema,
    images: z
        .array(ScannedImageSchema)
        .nullish()
        .transform(value => value ?? []),
});

const FindProjectSchema = z.object({
    data: z.object({
        enabled: z.boolean().catch(false),
        apps: z
            .array(ScannedAppSchema)
            .nullish()
            .transform(value => value ?? []),
    }),
    meta: BaseMetaApiSchema.nullish(),
});

const FindOneSchema = z.object({
    data: AppImageScanSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const ScanSchema = z.object({
    data: z.object({ task: z.object({ id: z.string() }) }),
    meta: BaseMetaApiSchema.nullish(),
});

const CreateIgnoreSchema = z.object({
    data: z.object({ id: z.string() }),
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

export class AppImageScanApiValidator {
    findOne = (response: ApiHttpResponse): AppImageScan_FindOne_Res => {
        return parseApiResponse({ response, schema: FindOneSchema });
    };

    findProject = (response: ApiHttpResponse): ProjectImageScan_Find_Res => {
        return parseApiResponse({ response, schema: FindProjectSchema });
    };

    scan = (response: ApiHttpResponse): AppImageScan_Scan_Res => {
        return parseApiResponse({ response, schema: ScanSchema });
    };

    createIgnore = (response: ApiHttpResponse): AppImageScan_CreateIgnore_Res => {
        return parseApiResponse({ response, schema: CreateIgnoreSchema });
    };

    deleteIgnore = (response: ApiHttpResponse): AppImageScan_DeleteIgnore_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
