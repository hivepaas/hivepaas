export const ENodeStatus = {
    Unknown: "unknown",
    Down: "down",
    Ready: "ready",
    Disconnected: "disconnected",
} as const;

export type ENodeStatus = (typeof ENodeStatus)[keyof typeof ENodeStatus];
