export const EGithubAppOwnerType = {
    Organization: "organization",
    User: "user",
} as const;

export type EGithubAppOwnerType = (typeof EGithubAppOwnerType)[keyof typeof EGithubAppOwnerType];
