export const ENodeStatus = {
    Ready: "ready",
    Pending: "pending",
    Disabled: "disabled",
} as const;

export type ENodeStatus = (typeof ENodeStatus)[keyof typeof ENodeStatus];
