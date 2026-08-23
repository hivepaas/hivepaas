import { z } from "zod";

import { isValidDomain } from "@application/shared/utils/domain";

export const MAX_HIVEPAAS_ROUTING_DOMAINS = 2;
export const MAX_HIVEPAAS_HTTP_DOMAINS = MAX_HIVEPAAS_ROUTING_DOMAINS;

export const RoutingSettingsRefSchema = z.object({
    id: z.string(),
    name: z.string(),
});
export const HttpSettingsRefSchema = RoutingSettingsRefSchema;

export const RoutingClientConfigSchema = z.object({
    enabled: z.boolean(),
    allowedIPs: z.string(),
});
export const HttpClientConfigSchema = RoutingClientConfigSchema;

export const RoutingRateLimitConfigSchema = z.object({
    enabled: z.boolean(),
    average: z.number().min(0),
    period: z.string(),
    burst: z.number().min(0),
    maxInFlightReq: z.number().min(0),
});
export const HttpRateLimitConfigSchema = RoutingRateLimitConfigSchema;

const DOMAIN_MAX_LEN = 100;

export const DomainFormSchema = z
    .object({
        enabled: z.boolean(),
        domain: z.string(),
        sslCert: RoutingSettingsRefSchema.optional(),
        clientConfig: RoutingClientConfigSchema,
        rateLimitConfig: RoutingRateLimitConfigSchema,
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

export const HivePaaSRoutingSettingsFormSchema = z.object({
    domains: z.array(DomainFormSchema).max(MAX_HIVEPAAS_ROUTING_DOMAINS, "Maximum 2 domains allowed"),
});
export const HivePaaSHttpSettingsFormSchema = HivePaaSRoutingSettingsFormSchema;

export type HivePaaSRoutingSettingsFormInput = z.input<typeof HivePaaSRoutingSettingsFormSchema>;
export type HivePaaSRoutingSettingsFormOutput = z.output<typeof HivePaaSRoutingSettingsFormSchema>;

export type HivePaaSHttpSettingsFormInput = HivePaaSRoutingSettingsFormInput;
export type HivePaaSHttpSettingsFormOutput = HivePaaSRoutingSettingsFormOutput;

export function createDefaultClientConfig(): z.infer<typeof RoutingClientConfigSchema> {
    return {
        enabled: false,
        allowedIPs: "",
    };
}

export function createDefaultRateLimitConfig(): z.infer<typeof RoutingRateLimitConfigSchema> {
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

export const emptyHivePaaSRoutingSettingsFormDefaults: HivePaaSRoutingSettingsFormInput = {
    domains: [],
};
export const emptyHivePaaSHttpSettingsFormDefaults = emptyHivePaaSRoutingSettingsFormDefaults;
