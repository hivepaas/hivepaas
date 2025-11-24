export const EProfileApiKeyAction = {
    Read: "read",
    Write: "write",
    Delete: "delete",
} as const;

export type EProfileApiKeyAction = (typeof EProfileApiKeyAction)[keyof typeof EProfileApiKeyAction];
