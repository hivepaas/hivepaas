export const EAccessTokenKind = {
    Github: "github",
    Gitlab: "gitlab",
    Gitea: "gitea",
    Bitbucket: "bitbucket",
    Gogs: "gogs",
} as const;

export type EAccessTokenKind = (typeof EAccessTokenKind)[keyof typeof EAccessTokenKind];
