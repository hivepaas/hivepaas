export const ESSHKeyKind = {
    Git: "git",
    Github: "github",
    Gitlab: "gitlab",
    Gitea: "gitea",
    Bitbucket: "bitbucket",
    Gogs: "gogs",
} as const;

export type ESSHKeyKind = (typeof ESSHKeyKind)[keyof typeof ESSHKeyKind];
