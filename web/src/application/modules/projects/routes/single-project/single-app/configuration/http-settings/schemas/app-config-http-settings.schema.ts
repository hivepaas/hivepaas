import { z } from "zod";
import { EHttpPathMode, ELBStrategy } from "~/projects/module-shared/enums";

import { isValidDomain } from "@application/shared/utils/domain";

export const HttpSettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

export const HttpBasicAuthConfigSchema = HttpSettingsRefSchema.extend({
    enabled: z.boolean(),
});

export const HttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    maxRequestBody: z.string(),
    memRequestBody: z.string(),
    allowedIPs: z.string(),
});

export const HttpHeaderConfigSchema = z.object({
    enabled: z.boolean(),
    autoContentType: z.boolean(),
    toAddToRequests: z.array(z.object({ key: z.string(), value: z.string() })),
    toRemoveFromRequests: z.array(z.object({ value: z.string() })),
    toAddToResponses: z.array(z.object({ key: z.string(), value: z.string() })),
    toRemoveFromResponses: z.array(z.object({ value: z.string() })),
});

export const HttpLBConfigSchema = z.object({
    strategy: z.union([z.literal(""), z.nativeEnum(ELBStrategy)]),
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

export const HttpPathRewriteConfigSchema = z.object({
    enabled: z.boolean(),
    prefixAdd: z.string(),
    prefixStrip: z.string(),
    prefixStripIsRegex: z.boolean(),
    pathReplace: z.string(),
    pathReplaceIsRegex: z.boolean(),
    pathReplaceWith: z.string(),
});

export const HttpCircuitBreakerConfigSchema = z.object({
    enabled: z.boolean(),
    expression: z.string(),
    checkPeriod: z.string(),
    fallbackDuration: z.string(),
    recoveryDuration: z.string(),
    responseCode: z.number().min(0),
});

export const HttpPathConfigSchema = z.object({
    enabled: z.boolean(),
    path: z.string().min(1, "Path is required"),
    mode: z.nativeEnum(EHttpPathMode),
    basicAuth: HttpBasicAuthConfigSchema.optional(),
    clientConfig: HttpClientConfigSchema.optional(),
    headerConfig: HttpHeaderConfigSchema.optional(),
    compressionConfig: HttpCompressionConfigSchema.optional(),
    rateLimitConfig: HttpRateLimitConfigSchema.optional(),
    pathRewriteConfig: HttpPathRewriteConfigSchema.optional(),
    circuitBreakerConfig: HttpCircuitBreakerConfigSchema.optional(),
});

const DOMAIN_MAX_LEN = 100; // mirrors backend base.DomainNameMaxLen

export const DomainFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string(),
        containerPort: z.number().int().min(1).max(65535),
        domainRedirect: z.string(),
        sslCert: HttpSettingsRefSchema.optional(),
        forceHttps: z.boolean(),
        basicAuth: HttpBasicAuthConfigSchema.optional(),
        lbConfig: HttpLBConfigSchema.optional(),
        clientConfig: HttpClientConfigSchema.optional(),
        headerConfig: HttpHeaderConfigSchema.optional(),
        compressionConfig: HttpCompressionConfigSchema.optional(),
        rateLimitConfig: HttpRateLimitConfigSchema.optional(),
        pathRewriteConfig: HttpPathRewriteConfigSchema.optional(),
        circuitBreakerConfig: HttpCircuitBreakerConfigSchema.optional(),
        paths: z.array(HttpPathConfigSchema),
    })
    .superRefine((values, ctx) => {
        const domain = values.domain.trim();
        if (!domain) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["domain"],
                message: "Domain is required",
            });
        } else if (!isValidDomain(domain, { maxLength: DOMAIN_MAX_LEN })) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["domain"],
                message: "Enter a valid domain (e.g. app.example.com)",
            });
        }

        const domainRedirect = values.domainRedirect.trim();
        if (domainRedirect && !isValidDomain(domainRedirect, { maxLength: DOMAIN_MAX_LEN })) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["domainRedirect"],
                message: "Enter a valid domain (e.g. other-domain.com)",
            });
        }
    });

export const AppConfigHttpSettingsFormSchema = z.object({
    exposePublicly: z.boolean(),
    domains: z.array(DomainFormSchema),
});

export type AppConfigHttpSettingsFormSchemaInput = z.input<typeof AppConfigHttpSettingsFormSchema>;
export type AppConfigHttpSettingsFormSchemaOutput = z.output<typeof AppConfigHttpSettingsFormSchema>;

export function createDefaultBasicAuthRef(): z.infer<typeof HttpBasicAuthConfigSchema> {
    return { id: "", name: "", enabled: true };
}

export function createDefaultLBConfig(): z.infer<typeof HttpLBConfigSchema> {
    return { strategy: "" };
}

export function createDefaultClientConfig(): z.infer<typeof HttpClientConfigSchema> {
    return {
        enabled: true,
        maxRequestBody: "",
        memRequestBody: "",
        allowedIPs: "",
    };
}

export function createDefaultHeaderConfig(): z.infer<typeof HttpHeaderConfigSchema> {
    return {
        enabled: true,
        autoContentType: false,
        toAddToRequests: [],
        toRemoveFromRequests: [],
        toAddToResponses: [],
        toRemoveFromResponses: [],
    };
}

export function createDefaultCompressionConfig(): z.infer<typeof HttpCompressionConfigSchema> {
    return {
        enabled: true,
        excludedContentTypes: "",
        includedContentTypes: "",
        minResponseBody: "1kb",
        defaultEncoding: "br",
    };
}

export function createDefaultRateLimitConfig(): z.infer<typeof HttpRateLimitConfigSchema> {
    return {
        enabled: true,
        average: 0,
        period: "",
        burst: 0,
        maxInFlightReq: 0,
    };
}

export function createDefaultPathRewriteConfig(): z.infer<typeof HttpPathRewriteConfigSchema> {
    return {
        enabled: true,
        prefixAdd: "",
        prefixStrip: "",
        prefixStripIsRegex: false,
        pathReplace: "",
        pathReplaceIsRegex: false,
        pathReplaceWith: "",
    };
}

export function createDefaultCircuitBreakerConfig(): z.infer<typeof HttpCircuitBreakerConfigSchema> {
    return {
        enabled: true,
        expression: "",
        checkPeriod: "",
        fallbackDuration: "",
        recoveryDuration: "",
        responseCode: 0,
    };
}

export const emptyDomain: z.input<typeof DomainFormSchema> = {
    enabled: true,
    domain: "",
    containerPort: 80,
    domainRedirect: "",
    forceHttps: true,
    lbConfig: createDefaultLBConfig(),
    paths: [],
};

export const emptyAppConfigHttpSettingsFormDefaults: AppConfigHttpSettingsFormSchemaInput = {
    exposePublicly: false,
    domains: [],
};
