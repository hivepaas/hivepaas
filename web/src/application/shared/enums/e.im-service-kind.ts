export const EImServiceKind = {
    Slack: "slack",
    Discord: "discord",
} as const;

export type EImServiceKind = (typeof EImServiceKind)[keyof typeof EImServiceKind];
