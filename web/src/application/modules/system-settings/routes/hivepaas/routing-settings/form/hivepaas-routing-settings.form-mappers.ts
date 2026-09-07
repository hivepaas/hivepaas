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
        sslCert: domain.sslCert?.id ? { id: domain.sslCert.id, name: domain.sslCert.name } : null,
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

/**
 * How long the dashboard asks to keep a routing change on trial.
 *
 * Longer than the server's own default, because a person at a browser is the
 * slowest caller it has: they are meant to go and look at the change before
 * vouching for it, and the confirm button is dead for the first ~25s while the
 * proxy catches up. Five minutes leaves a real margin for that, and the cost of
 * the extra time is a longer wait to get back into the dashboard - only the
 * dashboard, since these settings govern the HivePaaS routers alone.
 */
const CONFIRM_WINDOW = "5m";

export function mapFormValuesToPayload(
    values: HivePaaSRoutingSettingsFormOutput,
    updateVer: number,
): HivePaaSRoutingSettingsUpdatePayload {
    return {
        updateVer,
        confirmWindow: CONFIRM_WINDOW,
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
