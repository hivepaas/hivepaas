export const ENodeState = {
    Unknown: "unknown",
    Down: "down",
    Ready: "ready",
    Disconnected: "disconnected",
} as const;

export type ENodeState = (typeof ENodeState)[keyof typeof ENodeState];
