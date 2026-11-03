export const ERoutingProtocol = {
    HTTP: "http",
    TCP: "tcp",
    UDP: "udp",
} as const;

export type ERoutingProtocol = (typeof ERoutingProtocol)[keyof typeof ERoutingProtocol];
