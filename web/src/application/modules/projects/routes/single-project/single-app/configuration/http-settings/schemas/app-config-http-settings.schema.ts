import { z } from "zod";
import { EHttpPathMode } from "~/projects/module-shared/enums";

export const HttpSettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

export const HttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    maxRequestBody: z.string(),
    memRequestBody: z.string(),
    allowedIPs: z.string(),
});

export const HttpHeaderConfigSchema = z.object({
    toAddToRequests: z.array(z.object({ key: z.string(), value: z.string() })),
    toRemoveFromRequests: z.array(z.object({ value: z.string() })),
    toAddToResponses: z.array(z.object({ key: z.string(), value: z.string() })),
    toRemoveFromResponses: z.array(z.object({ value: z.string() })),
});

export const HttpCompressionConfigSchema = z.object({
    enabled: z.boolean(),
    excludedContentTypes: z.string(),
    includedContentTypes: z.string(),
    minResponseBody: z.string(),
    defaultEncoding: z.string(),
});

export const HttpRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number().min(0),
    period: z.string(),
    burst: z.number().min(0),
    maxInFlightReq: z.number().min(0),
});

export const HttpPathConfigSchema = z.object({
    path: z.string().min(1, "Path is required"),
    mode: z.nativeEnum(EHttpPathMode),
    basicAuth: HttpSettingsRefSchema.optional(),
    clientConfig: HttpClientConfigSchema.optional(),
    rateLimitConfig: HttpRateLimitConfigSchema.optional(),
});

export const DomainFormSchema = z.object({
    enabled: z.boolean(),
    domain: z.string(),
    containerPort: z.number().int().min(1).max(65535),
    domainRedirect: z.string(),
    sslCert: HttpSettingsRefSchema.optional(),
    forceHttps: z.boolean(),
    basicAuth: HttpSettingsRefSchema.optional(),
    clientConfig: HttpClientConfigSchema.optional(),
    headerConfig: HttpHeaderConfigSchema.optional(),
    compressionConfig: HttpCompressionConfigSchema.optional(),
    rateLimitConfig: HttpRateLimitConfigSchema.optional(),
    paths: z.array(HttpPathConfigSchema),
});

export const AppConfigHttpSettingsFormSchema = z.object({
    exposePublicly: z.boolean(),
    domains: z.array(DomainFormSchema),
});

export type AppConfigHttpSettingsFormSchemaInput = z.input<typeof AppConfigHttpSettingsFormSchema>;
export type AppConfigHttpSettingsFormSchemaOutput = z.output<typeof AppConfigHttpSettingsFormSchema>;

export function createDefaultBasicAuthRef(): z.infer<typeof HttpSettingsRefSchema> {
    return { id: "", name: "" };
}

export function createDefaultClientConfig(): z.infer<typeof HttpClientConfigSchema> {
    return {
        enabled: false,
        maxRequestBody: "",
        memRequestBody: "",
        allowedIPs: "",
    };
}

export function createDefaultHeaderConfig(): z.infer<typeof HttpHeaderConfigSchema> {
    return {
        toAddToRequests: [],
        toRemoveFromRequests: [],
        toAddToResponses: [],
        toRemoveFromResponses: [],
    };
}

export function createDefaultCompressionConfig(): z.infer<typeof HttpCompressionConfigSchema> {
    return {
        enabled: false,
        excludedContentTypes: "",
        includedContentTypes: "",
        minResponseBody: "",
        defaultEncoding: "",
    };
}

export function createDefaultRateLimitConfig(): z.infer<typeof HttpRateLimitConfigSchema> {
    return {
        enabled: false,
        average: 0,
        period: "",
        burst: 0,
        maxInFlightReq: 0,
    };
}

export const emptyDomain: z.input<typeof DomainFormSchema> = {
    enabled: true,
    domain: "",
    containerPort: 80,
    domainRedirect: "",
    forceHttps: false,
    paths: [],
};

export const emptyAppConfigHttpSettingsFormDefaults: AppConfigHttpSettingsFormSchemaInput = {
    exposePublicly: false,
    domains: [],
};
