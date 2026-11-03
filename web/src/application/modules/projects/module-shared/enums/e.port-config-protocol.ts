export const EPortConfigProtocol = {
    TCP: "tcp",
    UDP: "udp",
    SCTP: "sctp",
} as const;

export type EPortConfigProtocol = (typeof EPortConfigProtocol)[keyof typeof EPortConfigProtocol];
