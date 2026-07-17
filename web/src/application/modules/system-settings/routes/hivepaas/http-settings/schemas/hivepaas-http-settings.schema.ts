import { z } from "zod";

import { isValidDomain } from "@application/shared/utils/domain";

export const MAX_HIVEPAAS_HTTP_DOMAINS = 2;

export const HttpSettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});

export const HttpClientConfigSchema = z.object({
    enabled: z.boolean(),
    allowedIPs: z.string(),
});

export const HttpRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number().min(0),
    period: z.string(),
    burst: z.number().min(0),
    maxInFlightReq: z.number().min(0),
});

const DOMAIN_MAX_LEN = 100;

export const DomainFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string(),
        sslCert: HttpSettingsRefSchema.optional(),
        clientConfig: HttpClientConfigSchema,
        rateLimitConfig: HttpRateLimitConfigSchema,
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
    });

export const HivePaaSHttpSettingsFormSchema = z.object({
    domains: z.array(DomainFormSchema).max(MAX_HIVEPAAS_HTTP_DOMAINS, "Maximum 2 domains allowed"),
});

export type HivePaaSHttpSettingsFormInput = z.input<typeof HivePaaSHttpSettingsFormSchema>;
export type HivePaaSHttpSettingsFormOutput = z.output<typeof HivePaaSHttpSettingsFormSchema>;

export function createDefaultClientConfig(): z.infer<typeof HttpClientConfigSchema> {
    return {
        enabled: false,
        allowedIPs: "",
    };
}

export function createDefaultRateLimitConfig(): z.infer<typeof HttpRateLimitConfigSchema> {
    return {
        enabled: false,
        average: 10,
        period: "1m",
        burst: 20,
        maxInFlightReq: 10,
    };
}

export const emptyDomain: z.input<typeof DomainFormSchema> = {
    enabled: true,
    domain: "",
    clientConfig: createDefaultClientConfig(),
    rateLimitConfig: createDefaultRateLimitConfig(),
};

export const emptyHivePaaSHttpSettingsFormDefaults: HivePaaSHttpSettingsFormInput = {
    domains: [],
};
