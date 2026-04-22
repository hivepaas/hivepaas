import { type AxiosResponse } from "axios";
import { z } from "zod";
import {
    type AppHttpClientConfig,
    type AppHttpCompressionConfig,
    type AppHttpDomain,
    type AppHttpHeaderConfig,
    type AppHttpPathConfig,
    type AppHttpRateLimitConfig,
} from "~/projects/domain";
import { EHttpPathMode } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import { type AppHttpSettings_FindOne_Res } from "./app-http-settings.api.contracts";

const SettingRefSchema = z
    .object({
        id: z.string(),
        name: z.string(),
    })
    .passthrough();

const HttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    maxRequestBody: z.string(),
    memRequestBody: z.string(),
    allowedIPs: z.array(z.string()).nullish(),
});

const HttpHeaderConfigSchema = z.object({
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

const HttpPathConfigSchema = z.object({
    path: z.string(),
    mode: z.nativeEnum(EHttpPathMode),
    basicAuth: SettingRefSchema.nullish(),
    clientConfig: HttpClientConfigSchema.nullish(),
    rateLimitConfig: HttpRateLimitConfigSchema.nullish(),
});

const DomainSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    domainRedirect: z.string().optional(),
    sslCert: SettingRefSchema.nullish(),
    containerPort: z.number(),
    forceHttps: z.boolean().optional(),
    basicAuth: SettingRefSchema.nullish(),
    clientConfig: HttpClientConfigSchema.nullish(),
    headerConfig: HttpHeaderConfigSchema.nullish(),
    compressionConfig: HttpCompressionConfigSchema.nullish(),
    rateLimitConfig: HttpRateLimitConfigSchema.nullish(),
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

function mapSettingRef(raw: z.infer<typeof SettingRefSchema> | null | undefined): { id: string; name: string } | null {
    if (raw == null) {
        return null;
    }
    return { id: raw.id, name: raw.name };
}

function mapPath(raw: z.infer<typeof HttpPathConfigSchema>): AppHttpPathConfig {
    return {
        path: raw.path,
        mode: raw.mode,
        basicAuth: mapSettingRef(raw.basicAuth ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
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
        basicAuth: mapSettingRef(raw.basicAuth ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        headerConfig: mapHeaderConfig(raw.headerConfig ?? undefined),
        compressionConfig: mapCompressionConfig(raw.compressionConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
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
