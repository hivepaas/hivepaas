export const EOAuthKind = {
    Github: "github",
    Gitlab: "gitlab",
    Gitea: "gitea",
    Google: "google",
    MicrosoftOnline: "microsoft-online",
    OpenIDConnect: "openid-connect",
} as const;

export type EOAuthKind = (typeof EOAuthKind)[keyof typeof EOAuthKind];
