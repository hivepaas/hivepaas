export const EAccessTokenKind = {
    Bitbucket: "bitbucket",
    Cloudflare: "cloudflare",
    Gitea: "gitea",
    Github: "github",
    Gitlab: "gitlab",
    Gogs: "gogs",
} as const;

export type EAccessTokenKind = (typeof EAccessTokenKind)[keyof typeof EAccessTokenKind];
