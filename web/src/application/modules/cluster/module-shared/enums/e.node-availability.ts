export const ENodeAvailability = {
    Active: "active",
    Pause: "pause",
    Drain: "drain",
} as const;

export type ENodeAvailability = (typeof ENodeAvailability)[keyof typeof ENodeAvailability];
