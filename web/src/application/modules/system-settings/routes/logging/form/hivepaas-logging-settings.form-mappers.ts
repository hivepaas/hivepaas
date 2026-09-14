import type { HivePaaSLoggingSettings_UpdateOnePayload } from "~/system-settings/api/services";
import type { HivePaaSLoggingEndpoint, HivePaaSLoggingSettings } from "~/system-settings/domain";

import type { HivePaaSLoggingSettingsFormInput, HivePaaSLoggingSettingsFormOutput } from "../schemas";

type EndpointForm = HivePaaSLoggingSettingsFormInput["ingest"];

export const emptyLoggingEndpointForm: EndpointForm = {
    url: "",
    authMode: "none",
    username: "",
    password: "",
    bearerToken: "",
    headers: [],
    tlsSkipVerify: false,
};

// Which mode the stored endpoint is already in. A bearer token outranks basic
// auth because that is the order the client tries them in.
function toAuthMode(ep?: HivePaaSLoggingEndpoint | null): EndpointForm["authMode"] {
    if (ep?.bearerToken) {
        return "bearer";
    }
    if (ep?.username) {
        return "basic";
    }
    return "none";
}

// Headers are a map on the wire and a list in the form: a list is what the
// editor can hold a half-typed row in, and what keeps rows from reordering.
function toHeaderForms(headers?: Record<string, string>): EndpointForm["headers"] {
    return Object.entries(headers ?? {}).map(([key, value]) => ({ key, value }));
}

// Secrets arrive masked and are kept masked: sending the placeholder back is
// how the server knows to keep the stored value.
function toEndpointForm(ep?: HivePaaSLoggingEndpoint | null): EndpointForm {
    return {
        url: ep?.url ?? "",
        authMode: toAuthMode(ep),
        username: ep?.username ?? "",
        password: ep?.password ?? "",
        bearerToken: ep?.bearerToken ?? "",
        headers: toHeaderForms(ep?.headers),
        tlsSkipVerify: ep?.tlsSkipVerify ?? false,
    };
}

// Only the selected mode's credentials go out. What is left out the server
// clears, so the mode on screen is the mode that ends up stored: switching
// away from bearer does not leave a token behind to be sent by a later change.
function toEndpoint(f: EndpointForm): HivePaaSLoggingEndpoint {
    const basic = f.authMode === "basic";
    const bearer = f.authMode === "bearer";
    const headers: Record<string, string> = {};
    for (const h of f.headers) {
        // A row with no name has nothing to send under, and the map would key
        // it as "" and drop whichever such row came first.
        if (h.key) {
            headers[h.key] = h.value;
        }
    }
    return {
        url: f.url,
        ...(basic && f.username ? { username: f.username } : {}),
        ...(basic && f.password ? { password: f.password } : {}),
        ...(bearer && f.bearerToken ? { bearerToken: f.bearerToken } : {}),
        ...(Object.keys(headers).length > 0 ? { headers } : {}),
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
        backendManaged: s?.backend ? s.backend.managed || !s.backend.ingest?.url : true,
        nodeId: s?.backend.victoriaLogs?.node?.id ?? "",
        volumeId: s?.backend.victoriaLogs?.volume?.id ?? "",
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

export function toLoggingPayload(
    v: HivePaaSLoggingSettingsFormOutput,
    updateVer: number = 0,
): HivePaaSLoggingSettings_UpdateOnePayload {
    return {
        updateVer,
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
                      nodeId: { id: v.nodeId },
                      volumeId: { id: v.volumeId },
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
