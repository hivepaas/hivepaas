import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    HivePaaSRoutingClientConfig,
    HivePaaSRoutingDomain,
    HivePaaSRoutingRateLimitConfig,
} from "~/system-settings/domain";
import type {
    HivePaaSRoutingClientConfigSchema,
    HivePaaSRoutingDomainSchema,
    HivePaaSRoutingRateLimitConfigSchema,
} from "~/system-settings/module-shared/schemas";
import { HivePaaSRoutingSettingsEntitySchema } from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSRoutingSettings_FindOne_Res,
    HivePaaSRoutingSettings_UpdateOne_Res,
} from "./hivepaas-routing-settings.api.contracts";

const FindOneSchema = z.object({
    data: HivePaaSRoutingSettingsEntitySchema,
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
    raw: z.infer<typeof HivePaaSRoutingClientConfigSchema> | null | undefined,
): HivePaaSRoutingClientConfig | null {
    if (raw == null) {
        return null;
    }
    return {
        enabled: raw.enabled,
        allowedIPs: raw.allowedIPs ?? [],
    };
}

function mapRateLimitConfig(
    raw: z.infer<typeof HivePaaSRoutingRateLimitConfigSchema> | null | undefined,
): HivePaaSRoutingRateLimitConfig | null {
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

function mapDomain(raw: z.infer<typeof HivePaaSRoutingDomainSchema>): HivePaaSRoutingDomain {
    return {
        enabled: raw.enabled,
        domain: raw.domain,
        sslCert: mapSettingRef(raw.sslCert ?? undefined),
        clientConfig: mapClientConfig(raw.clientConfig ?? undefined),
        rateLimitConfig: mapRateLimitConfig(raw.rateLimitConfig ?? undefined),
    };
}

export class HivePaaSRoutingSettingsApiValidator {
    findOne = (response: AxiosResponse): HivePaaSRoutingSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                domains: data.domains?.map(mapDomain) ?? [],
                updateVer: data.updateVer,
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): HivePaaSRoutingSettings_UpdateOne_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };
}

export { HivePaaSRoutingSettingsApiValidator as HivePaaSHttpSettingsApiValidator };
