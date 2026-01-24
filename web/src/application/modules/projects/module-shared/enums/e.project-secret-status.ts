export const EProjectSecretStatus = {
    Active: "active",
    Pending: "pending",
    Disabled: "disabled",
    Expired: "expired",
} as const;

export type EProjectSecretStatus = (typeof EProjectSecretStatus)[keyof typeof EProjectSecretStatus];
