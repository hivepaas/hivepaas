export const EProfileApiKeyStatus = {
    Active: "active",
    Disabled: "disabled",
    Expired: "expired",
} as const;

export type EProfileApiKeyStatus = (typeof EProfileApiKeyStatus)[keyof typeof EProfileApiKeyStatus];
