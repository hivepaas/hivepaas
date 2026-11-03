export const EProfileApiKeyStatus = {
    Active: "active",
    Disabled: "disabled",
    Expired: "expired",
    Missing: "missing",
} as const;

export type EProfileApiKeyStatus = (typeof EProfileApiKeyStatus)[keyof typeof EProfileApiKeyStatus];
