import type { HivePaaSLoggingEndpoint, HivePaaSLoggingSettings } from "~/system-settings/domain";

import type { HivePaaSLoggingSettingsFormInput, HivePaaSLoggingSettingsFormOutput } from "../schemas";

type EndpointForm = HivePaaSLoggingSettingsFormInput["ingest"];

export const emptyLoggingEndpointForm: EndpointForm = {
    url: "",
    username: "",
    password: "",
    bearerToken: "",
    tlsSkipVerify: false,
};

// Secrets arrive masked and are kept masked: sending the placeholder back is
// how the server knows to keep the stored value.
function toEndpointForm(ep?: HivePaaSLoggingEndpoint | null): EndpointForm {
    return {
        url: ep?.url ?? "",
        username: ep?.username ?? "",
        password: ep?.password ?? "",
        bearerToken: ep?.bearerToken ?? "",
        tlsSkipVerify: ep?.tlsSkipVerify ?? false,
    };
}

function toEndpoint(f: EndpointForm): HivePaaSLoggingEndpoint {
    return {
        url: f.url,
        ...(f.username ? { username: f.username } : {}),
        ...(f.password ? { password: f.password } : {}),
        ...(f.bearerToken ? { bearerToken: f.bearerToken } : {}),
        tlsSkipVerify: f.tlsSkipVerify,
    };
}

export function toLoggingFormInput(s?: HivePaaSLoggingSettings): HivePaaSLoggingSettingsFormInput {
    return {
        enabled: s?.enabled ?? false,
        // One switch: container logs are collected as a set, so whichever of
        // the two the server holds means the same thing here.
        sources: {
            apps: (s?.sources.apps ?? true) || (s?.sources.hivepaas ?? false),
            hivepaas: s?.sources.hivepaas ?? false,
            traefikAccess: false,
            nodes: false,
        },
        backendManaged: s?.backend.managed ?? true,
        nodeId: s?.backend.victoriaLogs?.nodeId ?? "",
        volumeId: s?.backend.victoriaLogs?.volumeId ?? "",
        volumeSubpath: s?.backend.victoriaLogs?.volumeSubpath ?? "",
        retention: s?.backend.victoriaLogs?.retention ?? "30d",
        maxDiskUsagePercent: s?.backend.victoriaLogs?.maxDiskUsagePercent ?? null,
        ingest: toEndpointForm(s?.backend.ingest),
        query: toEndpointForm(s?.backend.query),
        forwards: (s?.forwards ?? []).map(f => ({
            ...toEndpointForm(f.endpoint),
            name: f.name,
            format: f.format ?? "jsonline",
        })),
    };
}

export function toLoggingPayload(v: HivePaaSLoggingSettingsFormOutput): HivePaaSLoggingSettings {
    return {
        enabled: v.enabled,
        // Only container logs are collected. The proxy's access log rides along
        // in traefik's own container output; the host's logs are not mounted.
        sources: { apps: v.sources.apps, hivepaas: v.sources.apps, traefikAccess: false, nodes: false },
        collector: { type: "vlagent", managed: true },
        backend: v.backendManaged
            ? {
                  type: "victoria-logs",
                  managed: true,
                  victoriaLogs: {
                      nodeId: v.nodeId,
                      volumeId: v.volumeId,
                      ...(v.volumeSubpath ? { volumeSubpath: v.volumeSubpath } : {}),
                      retention: v.retention,
                      ...(v.maxDiskUsagePercent ? { maxDiskUsagePercent: v.maxDiskUsagePercent } : {}),
                  },
              }
            : {
                  type: "victoria-logs",
                  managed: false,
                  ingest: toEndpoint(v.ingest),
                  ...(v.query.url ? { query: toEndpoint(v.query) } : {}),
              },
        forwards: v.forwards.map(f => ({ name: f.name, format: f.format, endpoint: toEndpoint(f) })),
    };
}
