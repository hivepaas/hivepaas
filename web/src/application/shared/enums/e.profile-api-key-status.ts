export const EProfileApiKeyStatus = {
    Active: "active",
    Disabled: "disabled",
} as const;

export type EProfileApiKeyStatus = (typeof EProfileApiKeyStatus)[keyof typeof EProfileApiKeyStatus];
