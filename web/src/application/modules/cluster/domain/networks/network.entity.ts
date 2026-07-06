import type { NetworkPublic } from "@application/shared/entities";

export type ClusterNetwork = NetworkPublic;

export interface ClusterNetworkCreatePayload {
    name: string;
    driver: string;
    enableIPv4: boolean;
    enableIPv6: boolean;
    internal: boolean;
    attachable: boolean;
    ingress: boolean;
    labels: Record<string, string>;
    options: Record<string, string>;
    availableInProjects: boolean;
    default?: boolean;
}

export interface ClusterNetworkUpdatePayload {
    updateVer: number;
    availableInProjects: boolean;
    default: boolean;
}

export interface ClusterNetworkUpdateStatusPayload extends ClusterNetworkUpdatePayload {
    status?: NetworkPublic["status"];
    expireAt?: Date | null;
}
