export const EMountConsistency = {
    Default: "default",
    Consistent: "consistent",
    Cached: "cached",
    Delegated: "delegated",
} as const;

export type EMountConsistency = (typeof EMountConsistency)[keyof typeof EMountConsistency];
