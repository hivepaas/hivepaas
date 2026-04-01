import { type AxiosResponse } from "axios";
import { z } from "zod";
import { EEndpointResolutionMode, EPortConfigProtocol, EPortConfigPublishMode } from "~/projects/module-shared/enums";

import { BaseMetaApiSchema, parseApiResponse } from "@infrastructure/api";

import {
    type AppNetworkSettings_FindOne_Res,
    type AppNetworkSettings_UpdateOne_Res,
} from "./app-network-settings.api.contracts";

const NetworkAttachmentSchema = z.object({
    id: z.string(),
    name: z.string().optional(),
    aliases: z.array(z.string()).nullish(),
});

const HostsFileEntrySchema = z.object({
    address: z.string().optional(),
    hostnames: z.array(z.string()).nullish(),
});

const DNSConfigSchema = z.object({
    nameservers: z.array(z.string()).nullish(),
    search: z.array(z.string()).nullish(),
    options: z.array(z.string()).nullish(),
});

const PortConfigSchema = z.object({
    target: z.number().optional(),
    published: z.number().optional(),
    protocol: z.nativeEnum(EPortConfigProtocol).optional(),
    publishMode: z.nativeEnum(EPortConfigPublishMode).optional(),
});

const EndpointSpecSchema = z.object({
    mode: z.nativeEnum(EEndpointResolutionMode).optional(),
    ports: z.array(PortConfigSchema).nullish(),
});

const AppNetworkSettingsSchema = z.object({
    networkAttachments: z.array(NetworkAttachmentSchema).nullish(),
    hostsFileEntries: z.array(HostsFileEntrySchema).nullish(),
    dnsConfig: DNSConfigSchema.nullish(),
    endpointSpec: EndpointSpecSchema.nullish(),
    updateVer: z.number(),
});

const FindOneSchema = z.object({
    data: AppNetworkSettingsSchema,
    meta: BaseMetaApiSchema.nullable(),
});

const UpdateOneSchema = z.object({
    data: z.object({ type: z.literal("success") }),
    meta: BaseMetaApiSchema.nullable(),
});

export class AppNetworkSettingsApiValidator {
    findOne = (response: AxiosResponse): AppNetworkSettings_FindOne_Res => {
        const { data, meta } = parseApiResponse({ response, schema: FindOneSchema });
        return {
            data: {
                dnsConfig: data.dnsConfig
                    ? {
                          options: data.dnsConfig.options ?? [],
                          nameservers: data.dnsConfig.nameservers ?? [],
                          search: data.dnsConfig.search ?? [],
                      }
                    : null,
                endpointSpec: data.endpointSpec
                    ? {
                          mode: data.endpointSpec.mode,
                          ports:
                              data.endpointSpec.ports?.map(item => ({
                                  target: item.target,
                                  published: item.published,
                                  protocol: item.protocol,
                                  publishMode: item.publishMode,
                              })) ?? [],
                      }
                    : null,
                updateVer: data.updateVer,
                networkAttachments:
                    data.networkAttachments?.map(item => ({
                        id: item.id,
                        name: item.name,
                        aliases: item.aliases ?? [],
                    })) ?? [],
                hostsFileEntries:
                    data.hostsFileEntries?.map(item => ({
                        address: item.address,
                        hostnames: item.hostnames ?? [],
                    })) ?? [],
            },
            meta,
        };
    };

    updateOne = (response: AxiosResponse): AppNetworkSettings_UpdateOne_Res => {
        return parseApiResponse({ response, schema: UpdateOneSchema });
    };
}
