import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type AppHttpClientConfig,
    type AppHttpCompressionConfig,
    type AppHttpDomain,
    type AppHttpHeaderConfig,
    type AppHttpLBConfig,
    type AppHttpPathConfig,
    type AppHttpPathRewriteConfig,
    type AppHttpRateLimitConfig,
} from "~/projects/domain";
import { EHttpPathMode, ELBStrategy } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { type AppHttpSettings_FindOne_Res } from "./app-http-settings.api.contracts";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .passthrough();

const BasicAuthConfigSchema = SettingRefSchema.extend({
    enabled: z.boolean().optional(),
});

const HttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    maxRequestBody: z.string(),
    memRequestBody: z.string(),
    allowedIPs: z.array(z.string()).nullish(),
});

const HttpHeaderConfigSchema = z.object({
    enabled: z.boolean().optional(),
    autoContentType: z.boolean().nullish(),
    toAddToRequests: z.record(z.string()).nullish(),
    toRemoveFromRequests: z.array(z.string()).nullish(),
    toAddToResponses: z.record(z.string()).nullish(),
    toRemoveFromResponses: z.array(z.string()).nullish(),
});

const HttpCompressionConfigSchema = z.object({
    enabled: z.boolean(),
    excludedContentTypes: z.array(z.string()).nullish(),
    includedContentTypes: z.array(z.string()).nullish(),
    minResponseBody: z.string(),
    defaultEncoding: z.string(),
});

const HttpRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number(),
    period: z.string(),
    burst: z.number(),
    maxInFlightReq: z.number(),
});

const HttpPathRewriteConfigSchema = z.object({
    enabled: z.boolean(),
    prefixAdd: z.string().nullish(),
    prefixStrip: z.string().nullish(),
    prefixStripIsRegex: z.boolean().nullish(),
    pathReplace: z.string().nullish(),
    pathReplaceIsRegex: z.boolean().nullish(),
    pathReplaceWith: z.string().nullish(),
});

const HttpLBConfigSchema = z.object({
    strategy: z.nativeEnum(ELBStrategy),
});

const HttpPathConfigSchema = z.object({
    enabled: z.boolean().optional(),
    path: z.string(),
    mode: z.nativeEnum(EHttpPathMode),
    basicAuth: BasicAuthConfigSchema.nullish(),
    clientConfig: HttpClientConfigSchema.nullish(),
    headerConfig: HttpHeaderConfigSchema.nullish(),
    compressionConfig: HttpCompressionConfigSchema.nullish(),
    rateLimitConfig: HttpRateLimitConfigSchema.nullish(),
    pathRewriteConfig: HttpPathRewriteConfigSchema.nullish(),
});

const DomainSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    domainRedirect: z.string().optional(),
    sslCert: SettingRefSchema.nullish(),
    containerPort: z.number(),
    forceHttps: z.boolean().optional(),
    basicAuth: BasicAuthConfigSchema.nullish(),
    lbConfig: HttpLBConfigSchema.nullish(),
    clientConfig: HttpClientConfigSchema.nullish(),
    headerConfig: HttpHeaderConfigSchema.nullish(),
    compressionConfig: HttpCompressionConfigSchema.nullish(),
    rateLimitConfig: HttpRateLimitConfigSchema.nullish(),
    pathRewriteConfig: HttpPathRewriteConfigSchema.nullish(),
    paths: z.array(HttpPathConfigSchema).nullish(),
});

const AppHttpSettingsSchema = z.object({
    internalEndpoints: z.array(z.string()).nullish(),
    domainSuggestion: z.string(),
    exposePublicly: z.boolean(),
    domains: z.array(DomainSchema).nullish(),
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppHttpSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

function mapClientConfig(raw: z.infer<typeof HttpClientConfigSchema> | null | undefined): AppHttpClientConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        maxRequestBody: raw.maxRequestBody,
        memRequestBody: raw.memRequestBody,
        allowedIPs: raw.allowedIPs ?? [],
    };
}

function mapHeaderConfig(raw: z.infer<typeof HttpHeaderConfigSchema> | null | undefined): AppHttpHeaderConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled ?? true,
        autoContentType: raw.autoContentType ?? false,
        toAddToRequests: raw.toAddToRequests ?? {},
        toRemoveFromRequests: raw.toRemoveFromRequests ?? [],
        toAddToResponses: raw.toAddToResponses ?? {},
        toRemoveFromResponses: raw.toRemoveFromResponses ?? [],
    };
}

function mapCompressionConfig(
    raw: z.infer<typeof HttpCompressionConfigSchema> | null | undefined,
): AppHttpCompressionConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        excludedContentTypes: raw.excludedContentTypes ?? [],
        includedContentTypes: raw.includedContentTypes ?? [],
        minResponseBody: raw.minResponseBody,
        defaultEncoding: raw.defaultEncoding,
    };
}

function mapRateLimitConfig(
    raw: z.infer<typeof HttpRateLimitConfigSchema> | null | undefined,
): AppHttpRateLimitConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        average: raw.average,
        period: raw.period,
        burst: raw.burst,
        maxInFlightReq: raw.maxInFlightReq,
    };
}

function mapPathRewriteConfig(
    raw: z.infer<typeof HttpPathRewriteConfigSchema> | null | undefined,
): AppHttpPathRewriteConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        prefixAdd: raw.prefixAdd ?? "",
        prefixStrip: raw.prefixStrip ?? "",
        prefixStripIsRegex: raw.prefixStripIsRegex ?? false,
        pathReplace: raw.pathReplace ?? "",
        pathReplaceIsRegex: raw.pathReplaceIsRegex ?? false,
        pathReplaceWith: raw.pathReplaceWith ?? "",
    };
}

function mapLBConfig(raw: z.infer<typeof HttpLBConfigSchema> | null | undefined): AppHttpLBConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        strategy: raw.strategy,
    };
}

function mapSettingRef(raw: z.infer<typeof SettingRefSchema> | null | undefined): { id: string; name: string } | null {
    if (raw == null) {
        return null;
    }
    return { id: raw.id, name: raw.name };
}

function mapBasicAuthConfig(
    raw: z.infer<typeof BasicAuthConfigSchema> | null | undefined,
): { id: string; name: string; enabled: boolean } | null {
    if (raw == null) {
        return null;
    }
    return { id: raw.id, name: raw.name, enabled: raw.enabled ?? true };
}

function mapPath(raw: z.infer<typeof HttpPathConfigSchema>): AppHttpPathConfig {
    return {
        enabled: raw.enabled ?? true,
        path: raw.path,
        mode: raw.mode,
        basicAuth: mapBasicAuthConfig(raw.basicAuth ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        headerConfig: mapHeaderConfig(raw.headerConfig ?? undefined),
        compressionConfig: mapCompressionConfig(raw.compressionConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
        pathRewriteConfig: mapPathRewriteConfig(raw.pathRewriteConfig ?? undefined),
    };
}

function mapDomain(raw: z.infer<typeof DomainSchema>): AppHttpDomain {
    return {
        enabled: raw.enabled,
        domain: raw.domain,
        domainRedirect: raw.domainRedirect,
        sslCert: mapSettingRef(raw.sslCert ?? undefined),
        containerPort: raw.containerPort,
        forceHttps: raw.forceHttps,
        basicAuth: mapBasicAuthConfig(raw.basicAuth ?? undefined),
        lbConfig: mapLBConfig(raw.lbConfig ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        headerConfig: mapHeaderConfig(raw.headerConfig ?? undefined),
        compressionConfig: mapCompressionConfig(raw.compressionConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
        pathRewriteConfig: mapPathRewriteConfig(raw.pathRewriteConfig ?? undefined),
        paths: raw.paths?.map(mapPath) ?? [],
    };
}

export class AppHttpSettingsApiValidator {
    findOne = (response: AxiosResponse): AppHttpSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                internalEndpoints: data.internalEndpoints ?? [],
                domainSuggestion: data.domainSuggestion,
                exposePublicly: data.exposePublicly,
                domains: data.domains?.map(mapDomain) ?? [],
                updateVer: data.updateVer,
            },
            meta,
        };
    };
}
