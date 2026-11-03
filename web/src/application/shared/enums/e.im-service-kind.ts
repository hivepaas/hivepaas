export const EImServiceKind = {
    Slack: "slack",
    Discord: "discord",
    Telegram: "telegram",
    Lark: "lark",
} as const;

export type EImServiceKind = (typeof EImServiceKind)[keyof typeof EImServiceKind];
