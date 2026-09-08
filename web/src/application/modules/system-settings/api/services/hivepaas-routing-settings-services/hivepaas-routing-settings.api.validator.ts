import type { AxiosResponse } from "axios";
import { z } from "zod";
import type {
    HivePaaSRoutingClientConfig,
    HivePaaSRoutingDomain,
    HivePaaSRoutingRateLimitConfig,
    SettingsPendingChange,
} from "~/system-settings/domain";
import type {
    HivePaaSRoutingClientConfigSchema,
    HivePaaSRoutingDomainSchema,
    HivePaaSRoutingRateLimitConfigSchema,
    SettingsPendingChangeSchema,
} from "~/system-settings/module-shared/schemas";
import {
    HivePaaSRoutingSettingsEntitySchema,
    SettingsPendingChangeSchema as PendingChangeSchema,
} from "~/system-settings/module-shared/schemas";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import type {
    HivePaaSRoutingSettings_ConfirmChange_Res,
    HivePaaSRoutingSettings_FindOne_Res,
    HivePaaSRoutingSettings_RevertChange_Res,
    HivePaaSRoutingSettings_UpdateOne_Res,
} from "./hivepaas-routing-settings.api.contracts";

const FindOneSchema = z.object({
    data: HivePaaSRoutingSettingsEntitySchema,
    meta: BaseMetaApiSchema.nullish(),
});

const MetaOnlySchema = z.object({
    meta: BaseMetaApiSchema.nullish(),
});

const UpdateOneSchema = z.object({
    data: PendingChangeSchema.nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

const RevertChangeSchema = z.object({
    data: z
        .object({
            reverted: z.boolean(),
            reason: z.string().nullish(),
        })
        .nullish(),
    meta: BaseMetaApiSchema.nullish(),
});

function mapPendingChange(
    raw: z.infer<typeof SettingsPendingChangeSchema> | null | undefined,
): SettingsPendingChange | null {
    if (raw == null) {
        return null;
    }
    return {
        changeId: raw.changeId,
        appliedAt: raw.appliedAt,
        confirmableFrom: raw.confirmableFrom,
        deadlineAt: raw.deadlineAt,
    };
}

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
                pendingChange: mapPendingChange(data.pendingChange ?? undefined),
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): HivePaaSRoutingSettings_UpdateOne_Res => {
        const { data } = parseApiResponse({ response, schema: UpdateOneSchema });
        return { data: { pendingChange: mapPendingChange(data ?? undefined) } };
    };

    confirmChange = (response: AxiosResponse): HivePaaSRoutingSettings_ConfirmChange_Res => {
        parseApiResponse({ response, schema: MetaOnlySchema });
        return { data: { type: "success" } };
    };

    revertChange = (response: AxiosResponse): HivePaaSRoutingSettings_RevertChange_Res => {
        const { data } = parseApiResponse({ response, schema: RevertChangeSchema });
        return { data: { reverted: data?.reverted ?? false, reason: data?.reason ?? null } };
    };
}

export { HivePaaSRoutingSettingsApiValidator as HivePaaSHttpSettingsApiValidator };
