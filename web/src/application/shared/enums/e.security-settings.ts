export const ESecuritySettings = {
    EnforceSSO: "enforce-sso",
    Password2FA: "password-2fa",
    PasswordOnly: "password-only",
} as const;

export type ESecuritySettings = (typeof ESecuritySettings)[keyof typeof ESecuritySettings];
