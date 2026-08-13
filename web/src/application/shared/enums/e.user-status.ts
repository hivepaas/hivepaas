export const EUserStatus = {
    Active: "active",
    Pending: "pending",
    Disabled: "disabled",
    Missing: "missing",
} as const;

export type EUserStatus = (typeof EUserStatus)[keyof typeof EUserStatus];
