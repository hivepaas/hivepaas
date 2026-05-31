export const ERepoWebhookKind = {
    Github: "github",
    Gitlab: "gitlab",
    Gitea: "gitea",
    Bitbucket: "bitbucket",
    Gogs: "gogs",
} as const;

export type ERepoWebhookKind = (typeof ERepoWebhookKind)[keyof typeof ERepoWebhookKind];
