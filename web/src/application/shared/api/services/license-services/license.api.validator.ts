import { z } from "zod";

import { type ApiHttpResponse, BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type { License_FindOne_Res, License_InstallKey_Res, License_Refresh_Res } from "./license.api.contracts";

/**
 * Status, plan, limit and refresh result stay strings rather than enums: a value
 * from a newer HivePaaS still parses, and is shown as it is - as the backend's
 * DisplayName intends - instead of failing the whole license.
 */
const text = z
    .string()
    .nullish()
    .transform(value => value ?? "");
const date = z.coerce
    .date()
    .nullish()
    .transform(value => value ?? null);

const LicenseSchema = z.object({
    status: z.string(),
    plan: z.string(),
    keyPlan: text,
    trial: z.boolean().catch(false),
    development: z.boolean().catch(false),
    licenseId: text,
    customerName: text,
    installationId: text,
    installations: z
        .array(z.string())
        .nullish()
        .transform(value => value ?? []),
    issuedAt: date,
    expiresAt: date,
    graceEndsAt: date,
    movedAt: date,
    reason: text,
    limits: z
        .array(
            z.object({
                limit: z.string(),
                max: z.number().int().nullable(),
                used: z.number().int().catch(0),
            }),
        )
        .nullish()
        .transform(value => value ?? []),
    portal: z
        .object({
            url: text,
            pricingUrl: text,
            trialUrl: text,
            accountUrl: text,
        })
        .nullish()
        .transform(value => value ?? null),
    refresh: z
        .object({
            enabled: z.boolean().catch(false),
            lastCheckedAt: date,
            lastResult: text,
            lastError: text,
        })
        .nullish()
        .transform(value => value ?? null),
});

/** GET answers with the license in force, and so does PUT, with the one the key put in force. */
const LicenseResponseSchema = z.object({
    data: LicenseSchema,
    meta: BaseMetaApiSchema.nullish(),
});

const RefreshSchema = z.object({
    data: z.object({
        result: z.string(),
        license: LicenseSchema,
    }),
    meta: BaseMetaApiSchema.nullish(),
});

export class LicenseApiValidator {
    findOne = (response: ApiHttpResponse): License_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: LicenseResponseSchema });
        return { data, meta };
    };

    installKey = (response: ApiHttpResponse): License_InstallKey_Res => {
        const { data, meta } = parseApiResponse({ response, schema: LicenseResponseSchema });
        return { data, meta };
    };

    refresh = (response: ApiHttpResponse): License_Refresh_Res => {
        const { data, meta } = parseApiResponse({ response, schema: RefreshSchema });
        return { data, meta };
    };
}
