export const ESignUpType = {
    PasswordOnly: "password-only",
    Password2FA: "password-2fa",
} as const;

export type ESignUpType = (typeof ESignUpType)[keyof typeof ESignUpType];
