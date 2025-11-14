export const EUserStatus = {
    Active: "active",
    Pending: "pending",
    Disabled: "disabled",
} as const;

export type EUserStatus = (typeof EUserStatus)[keyof typeof EUserStatus];
