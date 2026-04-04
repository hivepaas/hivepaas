import { type AppNetworkSettings } from "~/projects/domain";
import { EEndpointResolutionMode, EPortConfigProtocol, EPortConfigPublishMode } from "~/projects/module-shared/enums";

import { type AppConfigNetworksFormSchemaInput } from "../schemas";

export function mapAppNetworkSettingsToFormInput(data: AppNetworkSettings): AppConfigNetworksFormSchemaInput {
    return {
        networkAttachments: data.networkAttachments.map(item => ({
            id: item.id,
            name: item.name ?? "",
            aliasesText: (item.aliases ?? []).join(" "),
        })),
        hostsFileEntries: data.hostsFileEntries.map(item => ({
            address: item.address ?? "",
            hostnamesText: (item.hostnames ?? []).join(" "),
        })),
        dnsConfig: {
            nameservers: (data.dnsConfig?.nameservers ?? []).map(value => ({ value })),
            search: (data.dnsConfig?.search ?? []).map(value => ({ value })),
            options: (data.dnsConfig?.options ?? []).map(value => ({ value })),
        },
        resolutionMode: data.endpointSpec?.mode ?? EEndpointResolutionMode.VIP,
        portConfigs: (data.endpointSpec?.ports ?? []).map(item => ({
            published: item.published != null ? String(item.published) : "",
            target: item.target != null ? String(item.target) : "",
            protocol: item.protocol ?? EPortConfigProtocol.TCP,
            publishMode: item.publishMode ?? EPortConfigPublishMode.Host,
        })),
    };
}
