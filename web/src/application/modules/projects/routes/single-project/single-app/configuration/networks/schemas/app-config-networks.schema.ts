import { z } from "zod";
import { EEndpointResolutionMode, EPortConfigProtocol, EPortConfigPublishMode } from "~/projects/module-shared/enums";

export const NetworksFormAttachmentSchema = z.object({
    id: z.string(),
    aliasesText: z.string(),
});

export const NetworksFormHostsEntrySchema = z.object({
    address: z.string(),
    hostnamesText: z.string(),
});

export const NetworksFormPortConfigSchema = z.object({
    published: z.string(),
    target: z.string(),
    protocol: z.nativeEnum(EPortConfigProtocol),
    publishMode: z.nativeEnum(EPortConfigPublishMode),
});

export const NetworksFormDnsValueSchema = z.object({
    value: z.string(),
});

export const AppConfigNetworksFormSchema = z.object({
    networkAttachments: z.array(NetworksFormAttachmentSchema),
    hostsFileEntries: z.array(NetworksFormHostsEntrySchema),
    dnsConfig: z.object({
        nameservers: z.array(NetworksFormDnsValueSchema),
        search: z.array(NetworksFormDnsValueSchema),
        options: z.array(NetworksFormDnsValueSchema),
    }),
    resolutionMode: z.nativeEnum(EEndpointResolutionMode),
    portConfigs: z.array(NetworksFormPortConfigSchema),
});

export type AppConfigNetworksFormSchemaInput = z.input<typeof AppConfigNetworksFormSchema>;
export type AppConfigNetworksFormSchemaOutput = z.output<typeof AppConfigNetworksFormSchema>;

export const emptyAppConfigNetworksFormDefaults: AppConfigNetworksFormSchemaInput = {
    networkAttachments: [],
    hostsFileEntries: [],
    dnsConfig: {
        nameservers: [],
        search: [],
        options: [],
    },
    resolutionMode: EEndpointResolutionMode.VIP,
    portConfigs: [],
};
