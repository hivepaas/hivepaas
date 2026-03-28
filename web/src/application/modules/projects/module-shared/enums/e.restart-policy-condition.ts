export const ERestartPolicyCondition = {
    None: "none",
    OnFailure: "on-failure",
    Any: "any",
} as const;

export type ERestartPolicyCondition = (typeof ERestartPolicyCondition)[keyof typeof ERestartPolicyCondition];
