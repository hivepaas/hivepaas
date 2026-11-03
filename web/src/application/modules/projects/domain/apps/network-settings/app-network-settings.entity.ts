import type {
    EEndpointResolutionMode,
    EPortConfigProtocol,
    EPortConfigPublishMode,
} from "~/projects/module-shared/enums";

export type AppNetworkSettings = {
    networkAttachments: NetworkAttachment[];
    hostsFileEntries: HostsFileEntry[];
    dnsConfig: DNSConfig | null;
    endpointSpec: EndpointSpec | null;
    updateVer: number;
};

export type NetworkAttachment = {
    id: string;
    name?: string;
    aliases?: string[];
};

export type HostsFileEntry = {
    address?: string;
    hostnames?: string[];
};

export type DNSConfig = {
    nameservers?: string[];
    search?: string[];
    options?: string[];
};

export type EndpointSpec = {
    mode?: EEndpointResolutionMode;
    ports?: PortConfig[];
};

export type PortConfig = {
    target?: number;
    published?: number;
    protocol?: EPortConfigProtocol;
    publishMode?: EPortConfigPublishMode;
};
