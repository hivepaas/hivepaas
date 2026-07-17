import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    HivePaaSHttpClientConfig,
    HivePaaSHttpDomain,
    HivePaaSHttpRateLimitConfig,
} from "~/system-settings/domain";
import type {
    HivePaaSHttpClientConfigSchema,
    HivePaaSHttpDomainSchema,
    HivePaaSHttpRateLimitConfigSchema,
} from "~/system-settings/module-shared/schemas";
import { HivePaaSHttpSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSHttpSettings_FindOne_Res,
    HivePaaSHttpSettings_UpdateOne_Res,
} from "./hivepaas-http-settings.api.contracts";

const FindOneSchema = z.object({
    data: HivePaaSHttpSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

function mapSettingRef(raw: { id: string; name: string } | null | undefined): { id: string; name: string } | null {
    if (raw == null) {
        return null;
    }
    return { id: raw.id, name: raw.name };
}

function mapClientConfig(
    raw: z.infer<typeof HivePaaSHttpClientConfigSchema> | null | undefined,
): HivePaaSHttpClientConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        allowedIPs: raw.allowedIPs ?? [],
    };
}

function mapRateLimitConfig(
    raw: z.infer<typeof HivePaaSHttpRateLimitConfigSchema> | null | undefined,
): HivePaaSHttpRateLimitConfig | null {
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

function mapDomain(raw: z.infer<typeof HivePaaSHttpDomainSchema>): HivePaaSHttpDomain {
    return {
        enabled: raw.enabled,
        domain: raw.domain,
        sslCert: mapSettingRef(raw.sslCert ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
    };
}

export class HivePaaSHttpSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSHttpSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                domains: data.domains?.map(mapDomain) ?? [],
                updateVer: data.updateVer,
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): HivePaaSHttpSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}
