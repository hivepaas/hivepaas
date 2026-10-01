import { z } from "zod";
import { EHttpPathMode, ELBStrategy, ERoutingProtocol } from "~/projects/module-shared/enums";

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
});

export const HttpRateLimitConfigSchema = z
    .object({
        enabled: z.boolean(),
        average: z.number().min(0).optional(),
        period: z.string(),
        burst: z.number().min(0).optional(),
        maxInFlightReq: z.number().min(0).optional(),
    })
    .superRefine((values, ctx) => {
        // Traefik builds no limit without an average or an in-flight amount: one
        // turned on with neither looks set and limits nothing. The server refuses
        // it too.
        if (values.enabled && !values.average && !values.maxInFlightReq) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                path: ["average"],
                message: "Set an average, or a maximum of in-flight requests",
            });
        }
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

export const HttpWebsocketConfigSchema = z.object({
    enabled: z.boolean(),
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
    websocketConfig: HttpWebsocketConfigSchema.optional(),
});

const DOMAIN_MAX_LEN = 100; // mirrors backend base.DomainNameMaxLen

// Mirror the backend's limits on a domain's extra ALPN protocols.
const ALPN_PROTOCOLS_MAX = 10;
const ALPN_PROTOCOL_MAX_LEN = 255;
const ALPN_PROTOCOL_PATTERN = /^[\x21-\x7e]+$/;

/** The protocols typed in, separated by commas or spaces. */
export function parseAlpnProtocols(value: string): string[] {
    return value.split(/[\s,]+/).filter(Boolean);
}

export const DomainFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string(),
        protocol: z.nativeEnum(ERoutingProtocol),
        containerPort: z.number().int().min(1).max(65535),
        overridePort: z.boolean().optional(),
        tlsPassthrough: z.boolean(),
        extraAlpnProtocols: z.string(),
        domainRedirect: z.string(),
        sslCert: HttpSettingsRefSchema.nullable().optional(),
        forceHttps: z.boolean(),
        basicAuth: HttpBasicAuthConfigSchema.optional(),
        lbConfig: HttpLBConfigSchema.optional(),
        clientConfig: HttpClientConfigSchema.optional(),
        headerConfig: HttpHeaderConfigSchema.optional(),
        compressionConfig: HttpCompressionConfigSchema.optional(),
        rateLimitConfig: HttpRateLimitConfigSchema.optional(),
        pathRewriteConfig: HttpPathRewriteConfigSchema.optional(),
        circuitBreakerConfig: HttpCircuitBreakerConfigSchema.optional(),
        websocketConfig: HttpWebsocketConfigSchema.optional(),
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

        const alpn = parseAlpnProtocols(values.extraAlpnProtocols);
        const alpnError =
            alpn.length > ALPN_PROTOCOLS_MAX
                ? `At most ${ALPN_PROTOCOLS_MAX} protocols`
                : alpn.find(p => p.length > ALPN_PROTOCOL_MAX_LEN || !ALPN_PROTOCOL_PATTERN.test(p))
                  ? "A protocol is up to 255 printable ASCII characters (e.g. x-amzn-mqtt-ca)"
                  : new Set(alpn).size !== alpn.length
                    ? "A protocol is listed twice"
                    : null;
        if (alpnError) {
            ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["extraAlpnProtocols"], message: alpnError });
        }
    });

export const AppConfigRoutingSettingsFormSchema = z.object({
    port: z.number().int().min(1).max(65535),
    exposePublicly: z.boolean(),
    domains: z.array(DomainFormSchema),
});

export const AppConfigHttpSettingsFormSchema = AppConfigRoutingSettingsFormSchema;

export type AppConfigRoutingSettingsFormSchemaInput = z.input<typeof AppConfigRoutingSettingsFormSchema>;
export type AppConfigRoutingSettingsFormSchemaOutput = z.output<typeof AppConfigRoutingSettingsFormSchema>;

export type AppConfigHttpSettingsFormSchemaInput = AppConfigRoutingSettingsFormSchemaInput;
export type AppConfigHttpSettingsFormSchemaOutput = AppConfigRoutingSettingsFormSchemaOutput;

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
    };
}

/** Left empty, so the fields show their placeholders: a number put in for the user would be taken for advice. */
export function createDefaultRateLimitConfig(): z.infer<typeof HttpRateLimitConfigSchema> {
    return {
        enabled: true,
        average: undefined,
        period: "",
        burst: undefined,
        maxInFlightReq: undefined,
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

export function createDefaultWebsocketConfig(): z.infer<typeof HttpWebsocketConfigSchema> {
    return {
        enabled: true,
    };
}

export const emptyDomain: z.input<typeof DomainFormSchema> = {
    enabled: true,
    domain: "",
    protocol: ERoutingProtocol.HTTP,
    containerPort: 0,
    overridePort: false,
    tlsPassthrough: false,
    extraAlpnProtocols: "",
    domainRedirect: "",
    forceHttps: true,
    lbConfig: createDefaultLBConfig(),
    paths: [],
};

export const emptyAppConfigRoutingSettingsFormDefaults: AppConfigRoutingSettingsFormSchemaInput = {
    port: 0,
    exposePublicly: false,
    domains: [],
};

export const emptyAppConfigHttpSettingsFormDefaults = emptyAppConfigRoutingSettingsFormDefaults;
