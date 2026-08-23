import type {
    HivePaaSRoutingDomain,
    HivePaaSRoutingSettings,
    HivePaaSRoutingSettingsUpdatePayload,
} from "~/system-settings/domain";

import {
    type HivePaaSRoutingSettingsFormInput,
    type HivePaaSRoutingSettingsFormOutput,
    createDefaultClientConfig,
    createDefaultRateLimitConfig,
} from "../schemas";

function mapDomainToFormInput(domain: HivePaaSRoutingDomain): HivePaaSRoutingSettingsFormInput["domains"][number] {
    return {
        enabled: domain.enabled,
        domain: domain.domain,
        sslCert: domain.sslCert?.id ? { id: domain.sslCert.id, name: domain.sslCert.name } : undefined,
        clientConfig: domain.clientConfig
            ? {
                  enabled: domain.clientConfig.enabled,
                  allowedIPs: domain.clientConfig.allowedIPs.join(","),
              }
            : createDefaultClientConfig(),
        rateLimitConfig: domain.rateLimitConfig
            ? {
                  enabled: domain.rateLimitConfig.enabled,
                  average: domain.rateLimitConfig.average,
                  period: domain.rateLimitConfig.period,
                  burst: domain.rateLimitConfig.burst,
                  maxInFlightReq: domain.rateLimitConfig.maxInFlightReq,
              }
            : createDefaultRateLimitConfig(),
    };
}

export function mapHivePaaSRoutingSettingsToFormInput(data: HivePaaSRoutingSettings): HivePaaSRoutingSettingsFormInput {
    return {
        domains: data.domains.map(mapDomainToFormInput),
    };
}
export const mapHivePaaSHttpSettingsToFormInput = mapHivePaaSRoutingSettingsToFormInput;

export function mapFormValuesToPayload(
    values: HivePaaSRoutingSettingsFormOutput,
    updateVer: number,
): HivePaaSRoutingSettingsUpdatePayload {
    return {
        updateVer,
        domains: values.domains.map(domain => ({
            enabled: domain.enabled,
            domain: domain.domain,
            sslCert: { id: domain.sslCert?.id ?? "" },
            clientConfig: {
                enabled: domain.clientConfig.enabled,
                allowedIPs: domain.clientConfig.allowedIPs
                    .replace(/\n/g, ",")
                    .split(",")
                    .map(s => s.trim())
                    .filter(Boolean),
            },
            rateLimitConfig: {
                enabled: domain.rateLimitConfig.enabled,
                average: domain.rateLimitConfig.average,
                period: domain.rateLimitConfig.period,
                burst: domain.rateLimitConfig.burst,
                maxInFlightReq: domain.rateLimitConfig.maxInFlightReq,
            },
        })),
    };
}
