export const EProjectSecretStatus = {
    Active: "active",
    Pending: "pending",
    Disabled: "disabled",
    Expired: "expired",
    Missing: "missing",
} as const;

export type EProjectSecretStatus = (typeof EProjectSecretStatus)[keyof typeof EProjectSecretStatus];
