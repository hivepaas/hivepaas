export const ESettingStatus = {
    Active: "active",
    Pending: "pending",
    Disabled: "disabled",
    Expired: "expired",
} as const;

export type ESettingStatus = (typeof ESettingStatus)[keyof typeof ESettingStatus];
